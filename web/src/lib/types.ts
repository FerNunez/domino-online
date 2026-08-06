// Mirrors shared/types/types.go, shared/messaging/events.go, proto/game.proto,
// proto/lobby.proto. Backend proto responses use `omitempty` on int32 fields,
// so a `{left:0,right:0}` tile can arrive as `{}` — always read fields with a
// `?? 0` fallback rather than assuming presence.

export interface Tile {
  left: number;
  right: number;
}

export type Side = "left" | "right";

export function tileToString(tile: Tile): string {
  return `${tile.left}-${tile.right}`;
}

export function parseTileString(raw: string): Tile {
  const [left, right] = raw.split("-").map(Number);
  return { left: left ?? 0, right: right ?? 0 };
}

export function sameTile(a: Tile, b: Tile): boolean {
  return (a.left === b.left && a.right === b.right) || (a.left === b.right && a.right === b.left);
}

// Raw gRPC-shaped response fields (proto `RoundResult` has no scores).
export interface RoundResult {
  winner_id?: string;
  reason?: "domino" | "blocked" | string;
  // Only present on events relayed from RabbitMQ (shared/messaging events use
  // types.RoundResult, which does carry scores); absent on the direct
  // game.play_tile_response / game.pass_turn_response echoed to the actor.
  scores?: Record<string, number>;
}

export interface PlayerModel {
  id: string;
  name: string;
  slot: number;
  isConnected: boolean;
}

// pbl.LobbyStatus (proto/lobby.proto) is a plain int32 enum on the wire — the
// gateway marshals proto structs with stdlib encoding/json (not protojson),
// which emits the numeric value, not the enum name, and omits it entirely
// when 0 (LOBBY_STATUS_WAITING) due to `omitempty`. Treat a missing status
// the same as WAITING.
export const LobbyStatus = {
  WAITING: 0,
  IN_GAME: 1,
  FINISHED: 2,
} as const;
export type LobbyStatus = (typeof LobbyStatus)[keyof typeof LobbyStatus];

export interface LobbySettings {
  maxScore: number;
  turnTimerSeconds: number;
}

export interface LobbyModel {
  id: string;
  hostId: string;
  players: PlayerModel[];
  maxPlayers: number;
  status: LobbyStatus;
  settings: LobbySettings;
}

// --- server -> client WS event payloads ---

export interface GameStartedData {
  playerOrder: string[];
  handsSize: Record<string, number>;
  currentTurn: string;
}

export interface HandDealtData {
  playerID: string;
  playerTiles: string[]; // "L-R" strings, see parseTileString
}

export interface MoveMadeData {
  userID: string;
  tile: Tile;
  side: Side;
  next_turn: string;
  round_result?: RoundResult;
}

export interface PlayerPassedData {
  userID: string;
  next_turn: string;
  round_result?: RoundResult;
}

// Direct gRPC-shaped echo to the acting player only.
export interface PlayTileResponseData {
  board?: Tile[];
  hand?: Tile[];
  roundResult?: RoundResult;
}

export interface PassTurnResponseData {
  roundResult?: RoundResult;
}

export interface PlayerConnectionData {
  userID: string;
  lobbyID: string;
}

export interface PlayerJoinedData {
  userID: string;
  displayName: string;
  playerCount: number;
  maxPlayers: number;
}
