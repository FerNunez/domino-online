package grpc_clients

import (
	"domino/shared/env"
	pb "domino/shared/proto/trip"
	"domino/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type tripServiceClient struct {
	Client pb.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*tripServiceClient, error) {
	tripServiceURL := env.GetString("TRIP_SERVICE_URL", "trip-service:9093")

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(tripServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &tripServiceClient{
		Client: pb.NewTripServiceClient(conn),
		conn:   conn,
	}, nil
}
func (t *tripServiceClient) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}
