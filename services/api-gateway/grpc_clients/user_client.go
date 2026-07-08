package grpc_clients

import (
	"rebu/shared/env"
	pbu "rebu/shared/proto/user"
	"rebu/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type userServiceClient struct {
	Client pbu.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient() (*userServiceClient, error) {
	userServiceURL := env.GetString("USER_SERVICE_URL", "user-service:9095")

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(userServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &userServiceClient{
		Client: pbu.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *userServiceClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
