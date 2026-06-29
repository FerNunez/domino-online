package repository

import (
	"context"
	"fmt"
	"rebu/services/trip-service/internal/domain"

	pbd "rebu/shared/proto/driver"
	pb "rebu/shared/proto/trip"
)

type InmemRepository struct {
	trips      map[string]*domain.TripModel
	ridesFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *InmemRepository {
	return &InmemRepository{
		trips:      make(map[string]*domain.TripModel),
		ridesFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *InmemRepository) CreateTrip(_ context.Context, t *domain.TripModel) (*domain.TripModel, error) {
	r.trips[t.ID.Hex()] = t
	return t, nil
}
func (r *InmemRepository) SaveRideFare(_ context.Context, f *domain.RideFareModel) error {
	r.ridesFares[f.ID.Hex()] = f
	return nil
}
func (r *InmemRepository) GetRideFareByID(_ context.Context, id string) (*domain.RideFareModel, error) {
	f, ok := r.ridesFares[id]
	if !ok {
		return nil, fmt.Errorf("ride fare not found: %s", id)
	}
	return f, nil
}
func (r *InmemRepository) GetTripByID(_ context.Context, id string) (*domain.TripModel, error) {
	t, ok := r.trips[id]
	if !ok {
		return nil, fmt.Errorf("trip not found: %s", id)
	}
	return t, nil
}
func (r *InmemRepository) UpdateTrip(_ context.Context, tripID string, status string, driver *pbd.Driver) error {
	t, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found: %s", tripID)
	}
	t.Status = status
	if driver != nil {
		t.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			ProfilePicture: driver.ProfilePicture,
			CarPlate:       driver.CarPlate,
		}
	}
	return nil
}
