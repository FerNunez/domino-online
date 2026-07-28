package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"domino/services/game-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type gameConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.GameService
}

func NewGameConsumer(rabbitmq *messaging.RabbitMQ, service domain.GameService) *gameConsumer {
	return &gameConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

// Listen consumes lobby broadcast events (GameStarted) and player commands
// (play tile / pass) directed at this service.
func (c *gameConsumer) Listen() error {
	if err := c.rabbitmq.ConsumeMessages(messaging.NotifyLobby, c.handleLobbyEvent); err != nil {
		return err
	}
	return c.rabbitmq.ConsumeMessages(messaging.NotifyGame, c.handleGameCmd)
}

func (c *gameConsumer) handleLobbyEvent(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.LobbyEvent
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.GameStartCmd:
		var payload messaging.GameStartCmd
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handleGameStarted(ctx, envelope.LobbyID, payload)
	default:
		log.Printf("game-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

func (c *gameConsumer) handleGameCmd(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.LobbyEvent
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.PlayTileCmd:
		var payload messaging.MoveChangedData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handlePlayTile(ctx, envelope.LobbyID, payload)
	case contracts.PassCmd:
		var payload messaging.PlayerPassedData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.handlePass(ctx, envelope.LobbyID, payload)
	default:
		log.Printf("game-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

// handleGameStarted deals
// 1) creates game for the lobby
// 2) publishes new broadcast game data
// 3) player's hand plus the starting player's turn.
func (c *gameConsumer) handleGameStarted(ctx context.Context, lobbyID string, payload messaging.GameStartCmd) error {
	game, err := c.service.CreateGame(ctx, lobbyID, payload.PlayersID)
	if err != nil {
		return fmt.Errorf("couldn't create game: %w", err)
	}
	fmt.Printf("gmae created: %v\n", game)

	// Create game /Start game message response payload
	handSize := make(map[string]int, len(game.PlayerOrder))
	for userID, tiles := range game.Hands {
		handSize[userID] = len(tiles)
	}
	payloadOut := messaging.GameStartedData{
		PlayerOrder: game.PlayerOrder,
		HandsSize:   handSize,
		CurrentTurn: game.CurrentTurn,
	}
	data, err := json.Marshal(payloadOut)
	if err != nil {
		return err
	}

	// Publish: Game Created/ Started
	if err := c.rabbitmq.PublishMessage(ctx, contracts.GameStarted, contracts.LobbyEvent{
		LobbyID:  lobbyID,
		Data:     data,
		TargetID: "",
	}); err != nil {
		return err
	}

	// Convert PlayerID to []Tile into []string
	handString := make(map[string][]string, len(game.Hands))
	for playerID, hand := range game.Hands {
		tiles := make([]string, len(hand))
		for idx, t := range hand {
			tiles[idx] = t.String()
		}
		handString[playerID] = tiles
	}

	// Publish: Deal Hand string
	for playerID, tiles := range handString {
		payloadOut := messaging.HandDeltData{
			PlayerTiles: tiles,
			PlayerID:    playerID,
		}
		data, err := json.Marshal(payloadOut)
		if err != nil {
			return err
		}
		if err := c.rabbitmq.PublishMessage(ctx, contracts.HandDealt, contracts.LobbyEvent{
			LobbyID:  lobbyID,
			Data:     data,
			TargetID: playerID, // target
		}); err != nil {
			return err
		}

	}

	return c.publishTurnChanged(ctx, lobbyID, game.CurrentTurn)
}

func (c *gameConsumer) handlePlayTile(ctx context.Context, lobbyID string, payload messaging.MoveChangedData) error {
	tile, err := domain.ParseTile(payload.Tile)
	if err != nil {
		return fmt.Errorf("invalid tile in play command: %w", err)
	}

	game, result, err := c.service.PlayTile(ctx, lobbyID, payload.UserID, tile, payload.Side)
	if err != nil {
		return fmt.Errorf("couldn't play tile: %w", err)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := c.rabbitmq.PublishMessage(ctx, contracts.MoveMade, contracts.LobbyEvent{
		LobbyID: lobbyID,
		Data:    data,
	}); err != nil {
		return err
	}

	return c.finishTurn(ctx, lobbyID, game, result)
}

func (c *gameConsumer) handlePass(ctx context.Context, lobbyID string, payload messaging.PlayerPassedData) error {
	game, result, err := c.service.PassTurn(ctx, lobbyID, payload.UserID)
	if err != nil {
		return fmt.Errorf("couldn't pass turn: %w", err)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := c.rabbitmq.PublishMessage(ctx, contracts.PlayerPassed, contracts.LobbyEvent{
		LobbyID: lobbyID,
		Data:    data,
	}); err != nil {
		return err
	}

	return c.finishTurn(ctx, lobbyID, game, result)
}

// finishTurn publishes the round result if the round just ended, otherwise
// announces whose turn is next.
func (c *gameConsumer) finishTurn(ctx context.Context, lobbyID string, game *domain.GameModel, result *domain.RoundResult) error {
	if result != nil {
		return c.publishGameEnded(ctx, lobbyID, *result)
	}
	return c.publishTurnChanged(ctx, lobbyID, game.CurrentTurn)
}

func (c *gameConsumer) publishTurnChanged(ctx context.Context, lobbyID, userID string) error {
	data, err := json.Marshal(messaging.TurnChangedData{UserID: userID})
	if err != nil {
		return err
	}
	return c.rabbitmq.PublishMessage(ctx, contracts.TurnChanged, contracts.LobbyEvent{
		LobbyID: lobbyID,
		Data:    data,
	})
}

func (c *gameConsumer) publishGameEnded(ctx context.Context, lobbyID string, result domain.RoundResult) error {
	data, err := json.Marshal(messaging.GameEndedData{
		WinnerID: result.WinnerID,
		Reason:   result.Reason,
		Scores:   result.Scores,
	})
	if err != nil {
		return err
	}
	return c.rabbitmq.PublishMessage(ctx, contracts.GameEnded, contracts.LobbyEvent{
		LobbyID: lobbyID,
		Data:    data,
	})
}
