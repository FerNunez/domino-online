package grpc_clients

import (
	"rebu/shared/env"
	pb "rebu/shared/proto/driver"
	"rebu/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type driverServiceClient struct {
	Client pb.DriverServiceClient
	conn   *grpc.ClientConn
}

func NewDriverServiceClient() (*driverServiceClient, error) {
	driverServiceURL := env.GetString("DRIVER_SERVICE_URL", "driver-service:9092")

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(driverServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &driverServiceClient{
		Client: pb.NewDriverServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *driverServiceClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
