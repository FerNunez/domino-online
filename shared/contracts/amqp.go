package contracts

import "encoding/json"

// AmqpMessage is the envelope for every message published to RabbitMQ
// OwnerID routes the message to the correct Websocket connection in the gateway
// Data is the raw JSON payload, ketp as bytes so eadch consumer can unmashal into own type
type AmqpMessage struct {
	OwnerID string `json:"ownerId"`
	Data    []byte `json:"data"` // NOTE: Array of bytes to unmarshal into own type struct
}

type LobbyEvent struct {
	LobbyID  string          `json:"lobbyID"`
	TargetID string          `json:"targetID,omitempty"`
	Data     json.RawMessage `json:"data"`
}

// EVENTS:
// *.events.* = something happens
// *.cmd.* =instruction directed at a specific actor
const (
	// Lobby events
	PlayerJoinedLobby = "lobby.player_joined"  // {userID, displayName, playerCount, maxPlayers}
	PlayerLeftLobby   = "lobby.player_left"    //  {userID, playerCount}
	GameStartCmd      = "lobby.cmd.game_start" //

	// Game events
	GameStarted  = "game.game_started"  //
	HandDealt    = "game.hand_dealt"    // targeted to userID, {tiles: [...]}
	TurnChanged  = "game.turn_changed"  // {userID}
	MoveMade     = "game.move_made"     // {userID, tile, side}
	PlayerPassed = "game.player_passed" // {userID}
	GameEnded    = "game.ended"         // {winnerID, reason, scores}

	// Game commands: directed from a player to the game-service, consumed off
	// the NotifyGame queue (see shared/messaging.NotifyGame)
	PlayTileCmd = "game.cmd.play_tile" // {userID, tile, side}
	PassCmd     = "game.cmd.pass"      // {userID}
)
