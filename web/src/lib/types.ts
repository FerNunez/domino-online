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

// Mirrors shared/types.RoundResult. Winner/pip counts are per-team, not
// per-player — domino is played 1&3 vs 2&4 (see slotToTeamID below).
export interface RoundResult {
  winnerTeamID?: string;
  reason?: "REASON_DOMINO" | "REASON_BLOCKED" | string;
  // Summed pip liability of each team's remaining hand when the round ended
  // (not points scored — see latestRoundOverPoints in useGameConnection.ts).
  pipCounts?: Record<string, number>;
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
  gameID: string;
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
  // Every player's remaining tiles at the moment the round ended (mirrors
  // shared/messaging.RoundOverData.PlayerHands) — the round is already
  // decided, so revealing everyone's hand is safe. Keyed by player id.
  // Absent on the reconnect snapshot path (GameStateSnapshotData has no
  // equivalent field) — treat missing the same as no reveal to animate.
  playerHands?: Record<string, Tile[]>;
}

// Broadcast once, immediately after the RoundOver event that ends the
// match. Durable, explicit game-over signal — carries the same terminal
// fields RoundOver already carries (gameState/teamWinner/gameScores); not a
// second place to derive different game-over logic from.
export interface GameOverData {
  lobbyID: string;
  gameID: string;
  gameState: "GAME_STATUS_IN_PROGRESS" | "GAME_STATUS_FINISHED";
  gameScores: Record<string, number>;
  teamWinner: string;
}

// Direct ack to the client that sent game.cmd.round; the actual round
// transition arrives separately via the RoundStarted broadcast.
export interface NextRoundResponseData {
  roundNumber: number;
}

// Sent directly (not via a broadcast) right after the WebSocket opens, if a
// game was already in progress for this lobby — lets a (re)connecting client
// seed full state instead of waiting on the next live event. See
// services/api-gateway/ws.go and shared/messaging.GameStateSnapshotData.
// `board` is structurally identical to BoardState (lib/board.ts) so it can
// be assigned directly with no replay — not imported from there to avoid a
// circular import (board.ts already imports from this file).
export interface GameStateSnapshotData {
  gameID: string;
  gameNumber: number;
  status: "GAME_STATUS_IN_PROGRESS" | "GAME_STATUS_FINISHED";
  teamScores: Record<string, number>;
  teamWinner?: string;
  goalScore: number;
  roundID: string;
  roundNumber: number;
  roundStatus: "ROUND_STATUS_DEALT" | "ROUND_STATUS_IN_PROGRESS" | "ROUND_STATUS_OVER";
  playerOrder: string[];
  currentTurn: string;
  board: { tiles: Tile[]; leftEnd: number; rightEnd: number };
  hand: Tile[];
  handSizes: Record<string, number>;
  roundResult?: RoundResult;
}

// Accumulated client-side from RoundOver events for the lifetime of this
// connection only — lost on refresh/reconnect. The backend now durably
// persists round history and exposes it via GET /games/{id}/history
// (see services/history-service), but the frontend doesn't fetch it yet.
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

// --- REST history endpoints (services/history-service, via api-gateway) ---
// proto/history.proto's fields are written camelCase (like lobby.proto's
// hostId/maxPlayers), so these match the wire JSON directly — see lib/api.ts,
// which fetches them with no translation layer, same as getLobby.

export interface GameSummary {
  gameId: string;
  lobbyId: string;
  finalScores: Record<string, number>;
  teamWinner?: string;
  gameState: "GAME_STATUS_IN_PROGRESS" | "GAME_STATUS_FINISHED" | string;
  createdAt: string; // RFC3339
}

// One row of GET /games/{id}/history. Distinct from RoundResult (the WS
// shape embedded in RoundOver) since this carries its own roundId/
// roundNumber/startingPlayerId rather than being scoped to a live event.
export interface GameRoundSummary {
  roundId: string;
  roundNumber: number;
  startingPlayerId: string;
  winnerTeamId?: string;
  reason?: string;
  scores: Record<string, number>;
}

// One row of GET /rounds/{id}/actions' `actions` array — enough, together
// with HistoryHand, to replay a round move by move.
export interface HistoryAction {
  actionNumber: number;
  playerId: string;
  actionType: "play" | "pass" | string;
  tile?: Tile;
  side?: string;
  resultingLeftEnd: number;
  resultingRightEnd: number;
}

// One row of GET /rounds/{id}/actions' `hands` array — the hand each player
// was dealt at the start of that round.
export interface HistoryHand {
  playerId: string;
  tiles: Tile[];
}
