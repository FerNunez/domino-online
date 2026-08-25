package events

import (
	"context"
	"encoding/json"

	"domino/services/game-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	pbg "domino/shared/proto/game"
	"domino/shared/types"
)

type GameEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewGameEventPublisher(rmq *messaging.RabbitMQ) *GameEventPublisher {
	return &GameEventPublisher{
		rabbitmq: rmq,
	}
}

func (p *GameEventPublisher) PublishMoveMade(ctx context.Context, req *pbg.PlayTileRequest, round *domain.RoundModel, roundResult *types.RoundResult) error {
	payload := messaging.MoveMadeData{
		UserID:            req.UserId,
		Tile:              toTile(req.Tile),
		Side:              types.Side(req.Side),
		NextTurn:          round.CurrentTurn,
		RoundResult:       roundResult,
		RoundID:           round.ID,
		ActionNumber:      round.ActionCount,
		ResultingLeftEnd:  round.Board.LeftEnd,
		ResultingRightEnd: round.Board.RightEnd,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish to MoveMade
	return p.rabbitmq.PublishMessage(ctx, contracts.PlayerMoveMade, &contracts.DominoEvent{
		LobbyID:  round.LobbyID,
		TargetID: "",
		Data:     data,
	})
}

func (p *GameEventPublisher) PublishTurnPassed(ctx context.Context, req *pbg.PassTurnRequest, game *domain.RoundModel, roundResult *types.RoundResult) error {
	payload := messaging.PlayerPassedData{
		UserID:       req.UserId,
		NextTurn:     game.CurrentTurn,
		RoundResult:  roundResult,
		RoundID:      game.ID,
		ActionNumber: game.ActionCount,
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

func (p *GameEventPublisher) PublishRoundOver(ctx context.Context, game *domain.GameModel, roundResult *types.RoundResult) error {
	round := game.CurrentRound
	// create event
	payload := messaging.RoundOverData{
		LobbyID:        round.LobbyID,
		GameID:         game.ID,
		RoundID:        round.ID,
		RoundNumber:    round.RoundNumber,
		StartingPlayer: round.StartingPlayer,
		PlayerOrder:    round.PlayerOrder,
		RoundResult:    *roundResult,
		ActionCount:    round.ActionCount,
		GameScore:      game.TeamScores,
		GameState:      string(game.Status),
		TeamWinner:     game.TeamWinner,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish to MoveMade
	return p.rabbitmq.PublishMessage(ctx, contracts.RoundOver, &contracts.DominoEvent{
		LobbyID:  round.LobbyID,
		TargetID: "",
		Data:     data,
	})
}

func (p *GameEventPublisher) PublishGameOver(ctx context.Context, game *domain.GameModel) error {
	payload := messaging.GameOverData{
		LobbyID:    game.LobbyID,
		GameID:     game.ID,
		GameState:  string(game.Status),
		GameScore:  game.TeamScores,
		TeamWinner: game.TeamWinner,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.GameEnded, &contracts.DominoEvent{
		LobbyID:  game.LobbyID,
		TargetID: "",
		Data:     data,
	})
}

func (p *GameEventPublisher) PublishRoundStarted(ctx context.Context, round *domain.RoundModel) error {
	payload := messaging.RoundStartedData{
		LobbyID:        round.LobbyID,
		RoundID:        round.ID,
		StartingPlayer: &round.CurrentTurn,
		RoundNumber:    round.RoundNumber,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.RoundStarted, &contracts.DominoEvent{
		LobbyID:  round.LobbyID,
		TargetID: "",
		Data:     data,
	})
}

func (p *GameEventPublisher) PublishHandsDealt(ctx context.Context, round *domain.RoundModel) error {
	for playerID, tiles := range round.Hands {
		payload := messaging.HandDeltData{
			RoundID:     round.ID,
			PlayerTiles: tiles,
			PlayerID:    playerID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if err := p.rabbitmq.PublishMessage(ctx, contracts.HandDealt, &contracts.DominoEvent{
			LobbyID:  round.LobbyID,
			Data:     data,
			TargetID: playerID, // target
		}); err != nil {
			return err
		}
	}
	return nil
}

// -- Helper
func toTile(tile *pbg.Tile) types.Tile {
	return types.Tile{
		Left:  int(tile.Left),
		Right: int(tile.Right),
	}
}
