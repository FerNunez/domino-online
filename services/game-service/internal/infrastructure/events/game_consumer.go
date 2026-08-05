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
	if err := c.rabbitmq.ConsumeMessages(messaging.NotifyLobby, c.handleDominoEvent); err != nil {
		return err
	}
	// TODO: To consume game? or rather to publish into game?
	//return c.rabbitmq.ConsumeMessages(messaging.NotifyGame, c.handleGameCmd)
	return nil
}

func (c *gameConsumer) handleDominoEvent(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.DominoEvent
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

// handleGameStarted deals
// 1) creates game for the lobby
// 2) publishes new broadcast game data
// 3) player's hand plus the starting player's turn.
func (c *gameConsumer) handleGameStarted(ctx context.Context, lobbyID string, payload messaging.GameStartCmd) error {
	game, err := c.service.CreateGame(ctx, lobbyID, payload.PlayersID)
	if err != nil {
		return fmt.Errorf("couldn't create game: %w", err)
	}
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
	if err := c.rabbitmq.PublishMessage(ctx, contracts.GameStarted, &contracts.DominoEvent{
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
		if err := c.rabbitmq.PublishMessage(ctx, contracts.HandDealt, &contracts.DominoEvent{
			LobbyID:  lobbyID,
			Data:     data,
			TargetID: playerID, // target
		}); err != nil {
			return err
		}

	}

	return nil
}
