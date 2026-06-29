package main

import (
	"context"
	pb "rebu/shared/proto/driver"

	"google.golang.org/grpc"
)

type driverGrpcHandler struct {
	pb.UnimplementedDriverServiceServer
	service *Service
}

// NewGrpcHandler creates a new driverGrpcHandler with the service and attaches it to the server
func NewGrpcHandler(s *grpc.Server, svc *Service) {
	h := driverGrpcHandler{
		UnimplementedDriverServiceServer: pb.UnimplementedDriverServiceServer{},
		service:                          svc,
	}
	pb.RegisterDriverServiceServer(s, &h)
}

func (h *driverGrpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.service.RegisterDriver(req.DriverID, req.PackageSlug)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterDriverResponse{
		Driver: driver,
	}, nil
}

func (h *driverGrpcHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	h.service.UnregisterDriver(req.DriverID)
	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{Id: req.GetDriverID()},
	}, nil
}
