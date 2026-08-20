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

export function sameTile(a: Tile, b: Tile): boolean {
  return (
    (a.left === b.left && a.right === b.right) ||
    (a.left === b.right && a.right === b.left)
  );
}

// Mirrors shared/types.RoundResult. Winner/scores are per-team, not
// per-player — domino is played 1&3 vs 2&4 (see slotToTeamID below).
export interface RoundResult {
  winnerTeamID?: string;
  reason?: "REASON_DOMINO" | "REASON_BLOCKED" | string;
  // Only present on events relayed from RabbitMQ (shared/messaging events use
  // types.RoundResult, which does carry scores); absent on the direct
  // game.play_tile_response / game.pass_turn_response echoed to the actor.
  scores?: Record<string, number>;
}

// Mirrors shared/types.SlotToTeamID: slots 1&3 are one team, 2&4 the other.
export function slotToTeamID(slot: number): string {
  return slot % 2 === 1 ? "TEAM_A" : "TEAM_B";
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
  scores: Record<string, number>;
}

// Broadcast when round 2+ begins (round 1 has no such event — it's implied
// by GameStarted). Hands for the new round follow separately via HandDealt.
export interface RoundStartedData {
  lobbyID: string;
  roundID: string;
  nextStartingPlayer?: string;
  roundNumber: number;
}

// Broadcast when a round ends (domino-out or blocked/4-pass). Mirrors
// messaging.RoundOverData. Note roundResult.scores is pip-liability of each
// team's remaining hand, not "points scored" — compute points scored from
// the delta between consecutive gameScores snapshots instead.
export interface RoundOverData {
  lobbyID: string;
  roundID: string;
  roundNumber: number;
  roundResult: RoundResult;
  gameScores: Record<string, number>;
  gameState: "GAME_STATUS_IN_PROGRESS" | "GAME_STATUS_FINISHED";
  // Backend has no omitempty on this field — it arrives as "" (falsy) rather
  // than being omitted when the match hasn't ended. Use a falsy check.
  teamWinner?: string;
}

// Direct ack to the client that sent game.cmd.round; the actual round
// transition arrives separately via the RoundStarted broadcast.
export interface NextRoundResponseData {
  roundNumber: number;
}

// Client-only — accumulated locally from RoundOver events, since the
// backend only persists the current round (no history endpoint exists).
export interface RoundHistoryEntry {
  roundNumber: number;
  roundID: string;
  roundResult: RoundResult;
  gameScoresAfter: Record<string, number>;
  pointsThisRound: Record<string, number>;
}

export interface HandDealtData {
  playerID: string;
  playerTiles: Tile[];
}

export interface MoveMadeData {
  userID: string;
  tile: Tile;
  side: Side;
  nextTurn: string;
  roundResult?: RoundResult;
}

export interface PlayerPassedData {
  userID: string;
  nextTurn: string;
  roundResult?: RoundResult;
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
