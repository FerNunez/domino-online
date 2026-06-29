package domain

import (
	"context"
	"rebu/services/trip-service/pkg/types"
	pbd "rebu/shared/proto/driver"
	pb "rebu/shared/proto/trip"
	sharedtypes "rebu/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// message Trip{
//   string id = 1;
//   RideFare selectedFare = 2;
//   Route route= 3;
//   string status =4;
//   string userID = 5;
//   TripDriver driver =6;
// }

// TrypModel is the document model for a trip
// It embeds the full RideFare (including the route) so a single GetTripByID fetch
// returns everything needed to display the trip
type TripModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string             `bson:"userID"`
	Status   string             `bson:"status"`
	RideFare *RideFareModel     `bson:"rideFare"`
	Driver   *pb.TripDriver     `bson:"driver"` // this is already a proto var
}

// ToProto converts the trip to the proto representation for grpc responses
func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id:           t.ID.Hex(),
		SelectedFare: t.RideFare.ToProto(),
		Route:        t.RideFare.Route.ToProto(),
		Status:       t.Status,
		UserID:       t.UserID,
		Driver:       t.Driver,
	}
}

// TripRepository is the persistance interface for the domain layer
// The service layer depends on this interce; the infrastructure layer implements it
type TripRepository interface {
	CreateTrip(ctx context.Context, t *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, f *RideFareModel) error
	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error
}

type TripService interface {
	CreateTrip(ctx context.Context, f *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *sharedtypes.Coordinate, useOsrmApi bool) (*types.OsrmApiResponse, error)
	EstimatePackagePriceWithRoute(ctx context.Context, route *types.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *types.OsrmApiResponse) ([]*RideFareModel, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error
}
