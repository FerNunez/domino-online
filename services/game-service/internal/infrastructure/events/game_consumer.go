package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"

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

// Listen consumes messages from FindAvailableDriversQueue and applies handle which
// then notifies user NoDriverFound or Notifies driver for a game request command
func (c *gameConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.NotifyLobby, c.handle)
}

func (c *gameConsumer) handle(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.LobbyEvent
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	var payload messaging.GameStartedData
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.GameStarted:
		return c.handleGameStarted(ctx, envelope.LobbyID, payload)
	default:
		log.Printf("game-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

// handleFindAndNotify  find available drivers with seletected fare slug
// if no available => publish NoDriverFound with game userID
// if avaialable => publish DriverCmdGameRequest with gameData
func (c *gameConsumer) handleGameStarted(ctx context.Context, lobbyID string, payload messaging.GameStartedData) error {

	game, err := c.service.CreateGame(ctx, lobbyID)
	if err != nil {
		return fmt.Errorf("couldn't create game: %v", err)
	}
	type GameModel struct {
		LobbyID     string
		PlayerTiles [][]string // [userSlotPosition][tiles]
		Head        []string
		Tail        []string
	}

	// Find starting player
	found := false
	foundSlot := 0
	for slotPos, playerTiles := range game.PlayerTiles {
		// FIX: Correct searcher
		if slices.Contains(playerTiles, "12") {
			found = true
			foundSlot = slotPos
		}
		if found {
			break
		}
	}
	if !found {
		panic("expected a player to hold max expected tile")
	}

	// converting from slotPos -> hand, to userID hand
	playerTiles := make(map[string][]string, len(game.PlayerTiles))
	for idx, tiles := range game.PlayerTiles {
		playerID := payload.PlayersID[idx]
		playerTiles[playerID] = tiles
	}

	payloadOut := messaging.HandsDeltData{
		PlayerTiles:      playerTiles,
		StartingPlayerID: payload.PlayersID[foundSlot],
	}
	data, err := json.Marshal(payloadOut)
	if err != nil {
		return err
	}
	return c.rabbitmq.PublishMessage(ctx, contracts.HandDealt, contracts.LobbyEvent{
		Type:     "broadcast",
		LobbyID:  lobbyID,
		TargetID: "",
		Data:     data,
	})
}
