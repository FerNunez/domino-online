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
	rabbitmq  *messaging.RabbitMQ
	service   domain.GameService
	publisher *GameEventPublisher
}

func NewGameConsumer(rabbitmq *messaging.RabbitMQ, service domain.GameService, publisher *GameEventPublisher) (*gameConsumer, error) {
	_, err := rabbitmq.DeclareQueueAndBind(messaging.GameQueue, []string{contracts.GameStartCmd}, messaging.DominoExchange)
	if err != nil {
		return nil, fmt.Errorf("couldn't bind to game queue %w", err)
	}

	return &gameConsumer{
		rabbitmq:  rabbitmq,
		service:   service,
		publisher: publisher,
	}, nil
}

// Listen consumes lobby broadcast events (GameStarted) and player commands
// (play tile / pass) directed at this service.
func (c *gameConsumer) Listen() error {
	if err := c.rabbitmq.ConsumeMessages(messaging.GameQueue, c.handleDominoEvent); err != nil {
		return err
	}
	// TODO: To consume game? or rather to publish into game?
	// return c.rabbitmq.ConsumeMessages(messaging.NotifyGame, c.handleGameCmd)
	return nil
}

func (c *gameConsumer) handleDominoEvent(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.DominoEvent
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.GameStartCmd:
		var cmd messaging.GameStartCmd // GameStart Command from LobbyService
		if err := json.Unmarshal(envelope.Data, &cmd); err != nil {
			return err
		}
		return c.handleGameStarted(ctx, cmd.GameID, envelope.LobbyID, cmd)
	default:
		log.Printf("game-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

// handleGameStarted deals
// 1) creates game for the lobby
// 2) publishes new broadcast game data
// 3) player's hand plus the starting player's turn.
func (c *gameConsumer) handleGameStarted(ctx context.Context, gameID, lobbyID string, payload messaging.GameStartCmd) error {
	game, err := c.service.CreateGameWithID(ctx, gameID, lobbyID, payload.PlayersID)
	if err != nil {
		return fmt.Errorf("couldn't create game: %w", err)
	}
	round := game.CurrentRound
	if round == nil {
		return fmt.Errorf("can't retrieve the current game: %w", err)
	}

	// compute map UserID -> hand size
	handSize := make(map[string]int, len(round.PlayerOrder))
	for userID, tiles := range round.Hands {
		handSize[userID] = len(tiles)
	}
	payloadOut := messaging.GameStartedData{
		GameID:      gameID,
		PlayerOrder: round.PlayerOrder,
		HandsSize:   handSize,
		CurrentTurn: round.CurrentTurn,
		Scores:      game.TeamScores,
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

	// Publish: Round Started + Deal Hand (same as every later round, via NextRound)
	if err := c.publisher.PublishRoundStarted(ctx, round); err != nil {
		return err
	}
	return c.publisher.PublishHandsDealt(ctx, round)
}
