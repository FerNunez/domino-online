package grpc_clients

import (
	"domino/shared/env"
	pbu "domino/shared/proto/game"
	"domino/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gameServiceClient struct {
	Client pbu.GameServiceClient
	conn   *grpc.ClientConn
}

func NewGameServiceClient() (*gameServiceClient, error) {
	gameServiceURL := env.GetString("GAME_SERVICE_URL", "game-service:9093")

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(gameServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &gameServiceClient{
		Client: pbu.NewGameServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *gameServiceClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
