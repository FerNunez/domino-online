package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"domino/services/history-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	"domino/shared/types"

	amqp "github.com/rabbitmq/amqp091-go"
)

type historyConsumer struct {
	rabbitmq *messaging.RabbitMQ
	repo     domain.HistoryRepository
}

// NewHistoryConsumer binds GameStoreQueue, with dead-lettering enabled, to
// the game-service broadcast events needed to build a durable history:
// every accepted move/pass (for replay) and round/game outcomes (for
// summaries + analysis)
func NewHistoryConsumer(rabbitmq *messaging.RabbitMQ, repo domain.HistoryRepository) (*historyConsumer, error) {
	routingKeys := []string{
		contracts.GameStarted,
		contracts.HandDealt,
		contracts.PlayerMoveMade,
		contracts.PlayerPassed,
		contracts.RoundOver,
		contracts.GameEnded,
	}
	if err := rabbitmq.DeclareAndBindQueueWithDLQ(messaging.GameStoreQueue, routingKeys, messaging.DominoExchange); err != nil {
		return nil, fmt.Errorf("couldn't bind to game store queue: %w", err)
	}

	return &historyConsumer{
		rabbitmq: rabbitmq,
		repo:     repo,
	}, nil
}

func (c *historyConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.GameStoreQueue, c.handleDominoEvent)
}

func (c *historyConsumer) handleDominoEvent(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.DominoEvent
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.GameStarted:
		var payload messaging.GameStartedData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handleGameStarted(ctx, payload)
	case contracts.HandDealt:
		var payload messaging.HandDeltData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handleHandDealt(ctx, payload)
	case contracts.PlayerMoveMade:
		var payload messaging.MoveMadeData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handleMoveMade(ctx, payload)
	case contracts.PlayerPassed:
		var payload messaging.PlayerPassedData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handlePlayerPassed(ctx, payload)
	case contracts.RoundOver:
		var payload messaging.RoundOverData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handleRoundOver(ctx, payload)
	case contracts.GameEnded:
		var payload messaging.GameOverData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handleGameEnded(ctx, payload)
	default:
		log.Printf("history-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

func (c *historyConsumer) handleGameStarted(ctx context.Context, payload messaging.GameStartedData) error {
	for idx, playerID := range payload.PlayerOrder {
		slot := idx + 1
		if err := c.repo.UpsertGamePlayer(ctx, domain.GamePlayer{
			GameID:   payload.GameID,
			PlayerID: playerID,
			TeamID:   types.SlotToTeamID(slot),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (c *historyConsumer) handleHandDealt(ctx context.Context, payload messaging.HandDeltData) error {
	return c.repo.UpsertHand(ctx, domain.Hand{
		RoundID:  payload.RoundID,
		PlayerID: payload.PlayerID,
		Tiles:    payload.PlayerTiles,
	})
}

func (c *historyConsumer) handleMoveMade(ctx context.Context, payload messaging.MoveMadeData) error {
	tileLeft, tileRight := payload.Tile.Left, payload.Tile.Right
	return c.repo.InsertAction(ctx, domain.Action{
		RoundID:           payload.RoundID,
		ActionNumber:      payload.ActionNumber,
		PlayerID:          payload.UserID,
		ActionType:        types.Play,
		TileLeft:          &tileLeft,
		TileRight:         &tileRight,
		Side:              payload.Side,
		ResultingLeftEnd:  &payload.ResultingLeftEnd,
		ResultingRightEnd: &payload.ResultingRightEnd,
	})
}

func (c *historyConsumer) handlePlayerPassed(ctx context.Context, payload messaging.PlayerPassedData) error {
	return c.repo.InsertAction(ctx, domain.Action{
		RoundID:      payload.RoundID,
		ActionNumber: payload.ActionNumber,
		PlayerID:     payload.UserID,
		ActionType:   types.Pass,
	})
}

func (c *historyConsumer) handleRoundOver(ctx context.Context, payload messaging.RoundOverData) error {
	if err := c.repo.UpsertRound(ctx, domain.Round{
		ID:               payload.RoundID,
		GameID:           payload.GameID,
		RoundNumber:      payload.RoundNumber,
		StartingPlayerID: payload.StartingPlayer,
		PlayerOrder:      payload.PlayerOrder,
		WinnerTeamID:     payload.RoundResult.WinnerTeamID,
		Reason:           payload.RoundResult.Reason,
		Scores:           teamScoresToStringMap(payload.RoundResult.Scores),
		ActionCount:      payload.ActionCount,
	}); err != nil {
		return err
	}

	// Index every player into this game so their match history can be
	// looked up later; idempotent, so redelivery across rounds is safe.
	for idx, playerID := range payload.PlayerOrder {
		slot := idx + 1
		if err := c.repo.UpsertGamePlayer(ctx, domain.GamePlayer{
			GameID:   payload.GameID,
			PlayerID: playerID,
			TeamID:   types.SlotToTeamID(slot),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (c *historyConsumer) handleGameEnded(ctx context.Context, payload messaging.GameOverData) error {
	return c.repo.UpsertGame(ctx, domain.Game{
		ID:          payload.GameID,
		LobbyID:     payload.LobbyID,
		FinalScores: teamScoresToStringMap(payload.GameScore),
		TeamWinner:  string(payload.TeamWinner),
		GameState:   payload.GameState,
	})
}

func teamScoresToStringMap(scores map[types.TeamID]int) map[string]int {
	out := make(map[string]int, len(scores))
	for team, score := range scores {
		out[string(team)] = score
	}
	return out
}
