package events

// import (
// 	"context"
// 	"domino/services/game-service/internal/domain"
// 	"domino/shared/contracts"
// 	"domino/shared/messaging"
// 	"encoding/json"
// )
//
// type GameEventPublisher struct {
// 	rabbitmq *messaging.RabbitMQ
// }
//
// func NewGameEventPublisher(rmq *messaging.RabbitMQ) *GameEventPublisher {
// 	return &GameEventPublisher{
// 		rabbitmq: rmq,
// 	}
// }
//
// // Publishes contracts.GameStarted with starting player ID
// func (p *GameEventPublisher) PublishHandDealt(ctx context.Context, game *domain.GameModel) error {
// 	payload := messaging.HandDeltData{}
// 	data, err := json.Marshal(payload)
// 	if err != nil {
// 		return err
// 	}
//
// 	return p.rabbitmq.PublishMessage(ctx, contracts.GameStarted, contracts.LobbyEvent{
// 		Type:     "Broadcast", // TODO: is this better to convert to EventType
// 		LobbyID:  game.ID,
// 		TargetID: "",
// 		Data:     data,
// 	})
// }

// TODO: PublishPauseGame?
