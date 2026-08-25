import { API_URL } from "./constants";
import { GameRoundSummary, GameSummary, HistoryAction, HistoryHand, LobbyModel, LobbyStatus } from "./types";

const GUEST_TOKEN_KEY = "domino_guest_token";

export function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return sessionStorage.getItem(GUEST_TOKEN_KEY);
}

function storeToken(token: string) {
  sessionStorage.setItem(GUEST_TOKEN_KEY, token);
}

function wsTokenKey(lobbyID: string) {
  return `domino_ws_token_${lobbyID}`;
}

export function getStoredWsToken(lobbyID: string): string | null {
  if (typeof window === "undefined") return null;
  return sessionStorage.getItem(wsTokenKey(lobbyID));
}

function storeWsToken(lobbyID: string, wsToken: string) {
  sessionStorage.setItem(wsTokenKey(lobbyID), wsToken);
}

interface APIResponse<T> {
  data: T;
}

// Carries the HTTP status alongside the message so callers can distinguish
// e.g. a 404 "not ready yet" (worth retrying) from a real failure.
export class ApiError extends Error {
  constructor(message: string, public status: number) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string
): Promise<T> {
  const authToken = token ?? getStoredToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string> | undefined),
  };
  if (authToken) headers["Authorization"] = `Bearer ${authToken}`;

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });
  if (!res.ok) {
    const text = await res.text();
    throw new ApiError(`${options.method ?? "GET"} ${path} failed (${res.status}): ${text}`, res.status);
  }
  const body: APIResponse<T> = await res.json();
  return body.data;
}

// Ensures a guest identity exists for this tab, minting one if needed.
export async function ensureGuestToken(): Promise<string> {
  const existing = getStoredToken();
  if (existing) return existing;

  const data = await request<{ token: string }>("/auth/guest", { method: "POST" });
  storeToken(data.token);
  return data.token;
}

export async function createLobby(maxPlayers: number): Promise<{ lobbyID: string; wsToken: string }> {
  const res = await request<{ lobbyID: string; wsToken: string }>("/lobbies", {
    method: "POST",
    body: JSON.stringify({ maxPlayers }),
  });
  storeWsToken(res.lobbyID, res.wsToken);
  return res;
}

export async function joinLobby(lobbyID: string): Promise<{ lobbyID: string; wsToken: string }> {
  const res = await request<{ lobbyID: string; wsToken: string }>(`/lobbies/${lobbyID}/join`, { method: "POST" });
  storeWsToken(res.lobbyID, res.wsToken);
  return res;
}

export async function getLobby(lobbyID: string): Promise<LobbyModel> {
  const lobby = await request<LobbyModel>(`/lobbies/${lobbyID}`, { method: "GET" });
  // Proto's zero-value enum (LOBBY_STATUS_WAITING = 0) is dropped by
  // encoding/json's `omitempty` on the wire — see services/api-gateway's use
  // of plain json.Marshal on proto structs. Missing means waiting.
  return { ...lobby, status: lobby.status ?? LobbyStatus.WAITING };
}

// StartLobbyResponse is empty on the wire (services/lobby-service/internal/infrastructure/grpc/grpc_handler.go
// returns &pbl.StartLobbyResponse{}) — poll getLobby / wait for game.game_started over WS for the result.
export async function startLobby(lobbyID: string): Promise<void> {
  await request(`/lobbies/${lobbyID}/start`, { method: "POST" });
}

// --- history-service reads ---
// proto/history.proto's fields are camelCase, so api-gateway's plain
// encoding/json marshaling already matches lib/types.ts on the wire — no
// mapping layer needed here, same as getLobby above.

// Backend defaults to the caller's 20 most recent games (history-service's
// GetPlayerGames) — api-gateway doesn't forward limit/offset query params
// yet, so pagination isn't available from here.
export async function getPlayerGames(): Promise<GameSummary[]> {
  const games = await request<GameSummary[]>("/players/me/games", { method: "GET" });
  return (games ?? []).sort((a, b) => (b.createdAt ?? "").localeCompare(a.createdAt ?? ""));
}

// No game-level fields (final scores, winner, created date) come back here —
// GetGameHistoryResponse only carries rounds. Callers that need those should
// pull the matching entry out of getPlayerGames() instead.
export async function getGameHistory(gameId: string): Promise<GameRoundSummary[]> {
  const rounds = await request<GameRoundSummary[]>(`/games/${gameId}/history`, { method: "GET" });
  return (rounds ?? []).sort((a, b) => a.roundNumber - b.roundNumber);
}

export async function getRoundActions(
  roundId: string
): Promise<{ actions: HistoryAction[]; hands: HistoryHand[] }> {
  const data = await request<{ actions?: HistoryAction[]; hands?: HistoryHand[] }>(
    `/rounds/${roundId}/actions`,
    { method: "GET" }
  );
  return {
    actions: (data.actions ?? []).sort((a, b) => a.actionNumber - b.actionNumber),
    hands: data.hands ?? [],
  };
}
