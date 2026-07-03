package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"rebu/services/trip-service/internal/domain"
	"rebu/services/trip-service/pkg/types"
	"rebu/shared/env"
	pbd "rebu/shared/proto/driver"
	pb "rebu/shared/proto/trip"
	sharedtypes "rebu/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

// NewService creates the service layer wired to the given repository
func NewService(repo domain.TripRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateTrip(ctx context.Context, f *domain.RideFareModel) (*domain.TripModel, error) {
	trip := domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   f.UserID,
		Status:   "pending",
		RideFare: f,
		Driver:   &pb.TripDriver{},
	}
	return s.repo.CreateTrip(ctx, &trip)
}

// GetRoute fetches a driving route from the OSRM Api.
// Set useOsrmApi = false to return a deterministic mock route (for tests)
func (s *service) GetRoute(ctx context.Context, pickup, destination *sharedtypes.Coordinate, useOsrmApi bool) (*types.OsrmApiResponse, error) {
	// Use deterministic mock route: straight line
	if !useOsrmApi {
		return &types.OsrmApiResponse{
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry struct {
					Coordinates [][]float64 `json:"coordinates"`
				} `json:"geometry"`
			}{
				{
					Distance: 5.0,
					Duration: 600.,
					Geometry: struct {
						Coordinates [][]float64 `json:"coordinates"`
					}{
						Coordinates: [][]float64{
							{pickup.Latitude, pickup.Longitude},
							{destination.Latitude, destination.Longitude},
						},
					},
				},
			},
		}, nil
	}

	baseURL := env.GetString("OSRM_API", "http://router.project-osrm.org")
	// OSRM expects longitude,latitudfe (GeoJSON order oposite of our coordinates strcut)
	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson", baseURL, pickup.Longitude, pickup.Latitude, destination.Longitude, destination.Latitude)
	log.Printf("Fetching route from OSRM: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("OSRM request failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading OSMR response: %v", err)
	}
	var result types.OsrmApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing OSRM response: %v", err)
	}

	if len(result.Routes) == 0 {
		return nil, fmt.Errorf("OSRM returned no routes for pickup %+v and destination %+v", pickup, destination)
	}

	fmt.Printf("len(result.Routes): %v", result.Routes)
	fmt.Printf("OsrmApiResponse: %v", result)
	return &result, nil
}

// EstimatePackagePriceWithRoute calculates a fare estimate for each vehicle cat given a route
// Returns unsaved models, so caller must call GenerateTripFares to persist them with IDs
func (s *service) EstimatePackagePriceWithRoute(ctx context.Context, route *types.OsrmApiResponse) []*domain.RideFareModel {
	bases := getBasesFares()
	result := make([]*domain.RideFareModel, len(bases))
	for i, bf := range bases {
		result[i] = estimateFareRoute(bf, route)
	}
	return result
}

// GenerateTripFares persists fare estiamtes and assigns them mongodb
// After this call each fare each fare has a stable ID the rider can reference in CreateTrip
func (s *service) GenerateTripFares(ctx context.Context, fares []*domain.RideFareModel, userID string, route *types.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	result := make([]*domain.RideFareModel, len(fares))
	for i, f := range fares {
		fare := &domain.RideFareModel{
			ID:                primitive.NewObjectID(),
			UserID:            userID,
			PackageSlut:       f.PackageSlut,
			TotalPriceInCents: f.TotalPriceInCents,
			Route:             route,
		}
		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("saving fare: %w", err)
		}
		result[i] = fare
	}

	return result, nil
}

// GetAndValidateFare retrieves a fare and checks that it belongs to UserID
// This prevents an user from using another user's fareID to create a trip
func (s *service) GetAndValidateFare(ctx context.Context, fareID, userID string) (*domain.RideFareModel, error) {
	fare, err := s.repo.GetRideFareByID(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("getting fare: %w", err)
	}
	if fare == nil {
		return nil, fmt.Errorf("fare does not exist")
	}
	if fare.UserID != userID {
		return nil, fmt.Errorf("fare does not belong to user")
	}
	return fare, nil
}
func (s *service) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	return s.repo.GetTripByID(ctx, id)
}
func (s *service) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error {
	return s.repo.UpdateTrip(ctx, tripID, status, driver)
}

// --- Pricing Helpers ---

// Estimation = base + dist*mult + time*mult
func estimateFareRoute(base *domain.RideFareModel, route *types.OsrmApiResponse) *domain.RideFareModel {
	cfg := types.DefaultPricingConfig()
	r := route.Routes[0]
	total := base.TotalPriceInCents + r.Distance*cfg.PricePerUnitOfDistance + r.Duration*cfg.PricePerMinute
	return &domain.RideFareModel{
		TotalPriceInCents: total,
		PackageSlut:       base.PackageSlut,
	}
}

func getBasesFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{PackageSlut: "suv", TotalPriceInCents: 200},
		{PackageSlut: "sedan", TotalPriceInCents: 350},
		{PackageSlut: "van", TotalPriceInCents: 400},
		{PackageSlut: "luxury", TotalPriceInCents: 1000},
	}
}
