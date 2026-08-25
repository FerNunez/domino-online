package grpc_clients

import (
	"domino/shared/env"
	pbh "domino/shared/proto/history"
	"domino/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type historyServiceClient struct {
	Client pbh.HistoryServiceClient
	conn   *grpc.ClientConn
}

func NewHistoryServiceClient() (*historyServiceClient, error) {
	historyServiceURL := env.GetString("HISTORY_SERVICE_URL", "history-service:9094")

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(historyServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &historyServiceClient{
		Client: pbh.NewHistoryServiceClient(conn),
		conn:   conn,
	}, nil
}
func (t *historyServiceClient) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}
