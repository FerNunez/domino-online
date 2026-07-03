package grpc_clients

import (
	"rebu/shared/env"
	pbl "rebu/shared/proto/lobby"
	"rebu/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type lobbyServiceClient struct {
	Client pbl.LobbyServiceClient
	conn   *grpc.ClientConn
}

func NewLobbyServiceClient() (*lobbyServiceClient, error) {
	lobbyServiceURL := env.GetString("LOBBY_SERVICE_URL", "lobby-service:9094")

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(lobbyServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &lobbyServiceClient{
		Client: pbl.NewLobbyServiceClient(conn),
		conn:   conn,
	}, nil
}
func (t *lobbyServiceClient) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}
