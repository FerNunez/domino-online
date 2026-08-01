package contracts

import "encoding/json"

// DominoEvent is the envelope for every message published to RabbitMQ
type DominoEvent struct {
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

	// Private event
	HandDealt = "game.hand_dealt" // targeted to userID, {tiles: [...]}

	// Game events broadcasted
	GameStarted    = "game.game_started"     // Should I add game ID and lobbyID
	PlayerPassed   = "game.player_passed"    // {userID}
	PlayerMoveMade = "game.player_move_made" // {userID, tile, side}	// should I add here the next user and the reuslt?
	GameEnded      = "game.ended"            // {winnerID, reason, scores}

	// GPRC commands from player/apiGateway-wS to services
	PlayTileCmd = "game.cmd.play_tile" // {userID, tile, side}
	PassTurnCmd = "game.cmd.pass"      // {userID}

	// GPRC response to players/apiGateway-wS from a service rpc response
	PlayTileResponse = "game.play_tile_response" // {Board, Hand, RoundResult} // uses grpc response btw
	PassTurnResponse = "game.pass_turn_response" // {RoundResult}

)
