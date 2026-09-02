// Mirrors shared/contracts/amqp.go's event-name constants. See
// services/api-gateway/ws.go for the client-command handling and
// shared/messaging/queue_consumer.go for how broadcast/targeted events are
// relayed to the client.
import {
  GameOverData,
  GameStartedData,
  GameStateSnapshotData,
  HandDealtData,
  MoveMadeData,
  NextRoundResponseData,
  PassTurnResponseData,
  PlayerConnectionData,
  PlayerJoinedData,
  PlayerPassedData,
  PlayTileResponseData,
  RoundOverData,
  RoundStartedData,
  Side,
  Tile,
} from "./types";

export enum GameEvents {
  GameStarted = "game.game_started",
  HandDealt = "game.hand_dealt",
  PlayerMoveMade = "game.player_move_made",
  PlayerPassed = "game.player_passed",
  PlayTileResponse = "game.play_tile_response",
  PassTurnResponse = "game.pass_turn_response",
  RoundStarted = "game.round_started",
  RoundOver = "game.round_over",
  GameEnded = "game.ended",
  NextRoundResponse = "game.next_round_response",
  PlayTileCmd = "game.cmd.play_tile",
  PassTurnCmd = "game.cmd.pass",
  NextRoundCmd = "game.cmd.round",
  GameStateSync = "game.state_sync",
}

// Relayed verbatim by the gateway from lobby.player_connected /
// lobby.player_disconnected (see services/api-gateway/ws.go); the gateway's
// own AMQP queue is bound to "lobby.*" so these already reach the browser
// today, they were just previously dropped by isValidServerWsMessage.
export enum LobbyEvents {
  PlayerConnected = "lobby.player_connected",
  PlayerDisconnected = "lobby.player_disconnected",
  PlayerJoined = "lobby.player_joined",
}

// Messages sent from the server to the client via the websocket.
export type ServerWsMessage =
  | { type: GameEvents.GameStarted; data: GameStartedData }
  | { type: GameEvents.HandDealt; data: HandDealtData }
  | { type: GameEvents.PlayerMoveMade; data: MoveMadeData }
  | { type: GameEvents.PlayerPassed; data: PlayerPassedData }
  | { type: GameEvents.PlayTileResponse; data: PlayTileResponseData }
  | { type: GameEvents.PassTurnResponse; data: PassTurnResponseData }
  | { type: GameEvents.RoundStarted; data: RoundStartedData }
  | { type: GameEvents.RoundOver; data: RoundOverData }
  | { type: GameEvents.GameEnded; data: GameOverData }
  | { type: GameEvents.NextRoundResponse; data: NextRoundResponseData }
  | { type: LobbyEvents.PlayerConnected; data: PlayerConnectionData }
  | { type: LobbyEvents.PlayerDisconnected; data: PlayerConnectionData }
  | { type: LobbyEvents.PlayerJoined; data: PlayerJoinedData }
  | { type: GameEvents.GameStateSync; data: GameStateSnapshotData };

// Messages sent from the client to the server via the websocket.
export type ClientWsMessage =
  | { type: GameEvents.PlayTileCmd; data: { tile: Tile; side: Side } }
  | { type: GameEvents.PassTurnCmd; data: Record<string, never> }
  | { type: GameEvents.NextRoundCmd; data: Record<string, never> };

export function isValidGameEvent(event: string): event is GameEvents {
  return Object.values(GameEvents).includes(event as GameEvents);
}

export function isValidLobbyEvent(event: string): event is LobbyEvents {
  return Object.values(LobbyEvents).includes(event as LobbyEvents);
}

export function isValidServerWsMessage(message: {
  type: string;
}): message is ServerWsMessage {
  return isValidGameEvent(message.type) || isValidLobbyEvent(message.type);
}
