package events

import (
	"context"
	"domino/services/game-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	pbg "domino/shared/proto/game"
	"domino/shared/types"
	"encoding/json"
)

type GameEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewGameEventPublisher(rmq *messaging.RabbitMQ) *GameEventPublisher {
	return &GameEventPublisher{
		rabbitmq: rmq,
	}
}

func (p *GameEventPublisher) PublishMoveMade(ctx context.Context, req *pbg.PlayTileRequest, game *domain.GameModel, roundResult *types.RoundResult) error {
	// create event
	payload := messaging.MoveMadeData{
		UserID:      req.UserId,
		Tile:        toTile(req.Tile),
		Side:        req.Side,
		NextTurn:    game.CurrentTurn,
		RoundResult: roundResult,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish to MoveMade
	return p.rabbitmq.PublishMessage(ctx, contracts.PlayerMoveMade, &contracts.DominoEvent{
		LobbyID:  game.LobbyID,
		TargetID: "",
		Data:     data,
	})
}

func (p *GameEventPublisher) PublishTurnChanged(ctx context.Context, req *pbg.PassTurnRequest, game *domain.GameModel, roundResult *types.RoundResult) error {
	// create event
	payload := messaging.PlayerPassedData{
		UserID:      req.UserId,
		NextTurn:    game.CurrentTurn,
		RoundResult: roundResult,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish to MoveMade
	return p.rabbitmq.PublishMessage(ctx, contracts.PlayerPassed, &contracts.DominoEvent{
		LobbyID:  game.LobbyID,
		TargetID: "",
		Data:     data,
	})
}

func toTile(tile *pbg.Tile) types.Tile {
	return types.Tile{
		Left:  int(tile.Left),
		Right: int(tile.Right),
	}
}
