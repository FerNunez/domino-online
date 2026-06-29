package grpc

import (
	"context"
	"log"
	"rebu/services/trip-service/internal/domain"
	"rebu/services/trip-service/internal/infrastructure/events"
	pb "rebu/shared/proto/trip"
	sharedtypes "rebu/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer

	service   domain.TripService
	publisher *events.TripEventPublisher // publishes in rabbitmq
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedTripServiceServer: pb.UnimplementedTripServiceServer{},
		service:                        service,
		publisher:                      publisher,
	}
	pb.RegisterTripServiceServer(server, h)
	return h
}

// PreviewTrip creates a Route and estimates its cost and generate ride fares
func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := sharedtypes.Coordinate{
		Latitude:  req.GetStartLocation().Latitude,
		Longitude: req.GetStartLocation().Longitude,
	}
	destination := sharedtypes.Coordinate{
		Latitude:  req.GetEndLocation().Latitude,
		Longitude: req.GetEndLocation().Longitude,
	}
	route, err := h.service.GetRoute(ctx, &pickup, &destination, true)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	estimated := h.service.EstimatePackagePriceWithRoute(ctx, route)

	fares, err := h.service.GenerateTripFares(ctx, estimated, req.GetUserID(), route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate fares: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

// CreateTrip gets and validates fare from user,
// and then create service from fare
// publishes an event trip created
// and return trip of ID
func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fare, err := h.service.GetAndValidateFare(ctx, req.RideFareID, req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate fare: %v", err)
	}
	trip, err := h.service.CreateTrip(ctx, fare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}

	// publishing 'trip.event.created'
	if err := h.publisher.PublishTripCreated(ctx, trip); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish trip created event: %v", err)
	}
	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}
