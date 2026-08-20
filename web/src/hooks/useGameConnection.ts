import { useEffect, useRef, useState } from "react";
import { WEBSOCKET_URL } from "@/lib/constants";
import {
  GameEvents,
  isValidServerWsMessage,
  LobbyEvents,
  ServerWsMessage,
} from "@/lib/contracts";
import { applyMove, BoardState, emptyBoard } from "@/lib/board";
import {
  PlayerModel,
  RoundHistoryEntry,
  RoundOverData,
  Side,
  Tile,
} from "@/lib/types";

const STARTING_HAND_SIZE = 7;
const NEXT_ROUND_TIMEOUT_MS = 5000;

interface GameConnectionState {
  connected: boolean;
  gameStarted: boolean;
  playerOrder: string[];
  currentTurn: string | null;
  hand: Tile[];
  board: BoardState;
  // Cumulative match score per team, updated on GameStarted/RoundOver.
  gameScores: Record<string, number>;
  roundNumber: number;
  gameStatus: "GAME_STATUS_IN_PROGRESS" | "GAME_STATUS_FINISHED";
  teamWinner: string | null;
  // Non-null only between a RoundOver event and the next RoundStarted/GameStarted.
  roundOver: RoundOverData | null;
  // Points scored by each team in the round that just ended (gameScores delta,
  // not roundOver.roundResult.scores, which is pip-liability, not points).
  latestRoundOverPoints: Record<string, number> | null;
  // Accumulated client-side only — the backend doesn't persist past rounds.
  roundHistory: RoundHistoryEntry[];
  nextRoundRequested: boolean;
  error: string | null;
  playerConnectivity: Record<string, boolean>;
  // Players who joined after this client's initial getLobby snapshot,
  // relayed live via lobby.player_joined. Merged onto that snapshot in
  // the lobby page rather than replacing it.
  joinedPlayers: PlayerModel[];
  // userID -> remaining tile count, for every seat (opponents included —
  // their actual tiles are private and never sent to other clients).
  handCounts: Record<string, number>;
}

const initialState: GameConnectionState = {
  connected: false,
  gameStarted: false,
  playerOrder: [],
  currentTurn: null,
  hand: [],
  board: emptyBoard,
  gameScores: {},
  roundNumber: 1,
  gameStatus: "GAME_STATUS_IN_PROGRESS",
  teamWinner: null,
  roundOver: null,
  latestRoundOverPoints: null,
  roundHistory: [],
  nextRoundRequested: false,
  error: null,
  playerConnectivity: {},
  joinedPlayers: [],
  handCounts: {},
};

// Mirrors web/src/hooks/useRiderStreamConnection.ts's shape: connect in a
// useEffect, switch incoming messages into state slices, guard sends on
// readyState. No reconnect/backoff for this first version.
export function useGameConnection(
  lobbyID: string | undefined,
  wsToken: string | undefined,
  userID: string | undefined,
) {
  const [state, setState] = useState<GameConnectionState>(initialState);
  const wsRef = useRef<WebSocket | null>(null);
  const nextRoundTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );

  const clearNextRoundTimeout = () => {
    if (nextRoundTimeoutRef.current) {
      clearTimeout(nextRoundTimeoutRef.current);
      nextRoundTimeoutRef.current = null;
    }
  };

  useEffect(() => {
    if (!lobbyID || !wsToken) return;

    const ws = new WebSocket(
      `${WEBSOCKET_URL}/lobbies/${lobbyID}/ws?wsToken=${wsToken}`,
    );
    wsRef.current = ws;

    ws.onopen = () => setState((s) => ({ ...s, connected: true }));
    ws.onclose = () => setState((s) => ({ ...s, connected: false }));
    ws.onerror = () =>
      setState((s) => ({ ...s, error: "WebSocket error occurred" }));

    ws.onmessage = (event) => {
      let message: ServerWsMessage;
      try {
        message = JSON.parse(event.data);
      } catch {
        return;
      }
      if (!isValidServerWsMessage(message)) return;

      switch (message.type) {
        case GameEvents.GameStarted: {
          const { playerOrder, currentTurn, handsSize, scores } =
            message.data;
          setState((s) => ({
            ...s,
            gameStarted: true,
            playerOrder,
            currentTurn,
            board: emptyBoard,
            handCounts: handsSize ?? {},
            gameScores: scores ?? { TEAM_A: 0, TEAM_B: 0 },
            roundNumber: 1,
            gameStatus: "GAME_STATUS_IN_PROGRESS",
            teamWinner: null,
            roundOver: null,
            latestRoundOverPoints: null,
            roundHistory: [],
            nextRoundRequested: false,
          }));
          break;
        }
        case GameEvents.HandDealt: {
          if (message.data.playerID !== userID) break;
          setState((s) => ({ ...s, hand: message.data.playerTiles }));
          break;
        }
        case GameEvents.PlayerMoveMade: {
          const { userID: actorID, tile, side, nextTurn } = message.data;
          // roundResult is no longer read here — RoundOver is now the sole
          // source of round-end state, so it can't go stale across rounds.
          setState((s) => ({
            ...s,
            board: applyMove(s.board, tile, side as Side),
            currentTurn: nextTurn,
            handCounts: {
              ...s.handCounts,
              [actorID]: Math.max(0, (s.handCounts[actorID] ?? 0) - 1),
            },
          }));
          break;
        }
        case GameEvents.PlayerPassed: {
          const { nextTurn } = message.data;
          setState((s) => ({ ...s, currentTurn: nextTurn }));
          break;
        }
        case GameEvents.PlayTileResponse: {
          // Board is derived from the broadcast PlayerMoveMade for every
          // client uniformly; only sync the actor's private hand here.
          if (message.data.hand) {
            setState((s) => ({ ...s, hand: message.data.hand as Tile[] }));
          }
          break;
        }
        case GameEvents.PassTurnResponse:
          break;
        case GameEvents.RoundStarted: {
          const { nextStartingPlayer, roundNumber } = message.data;
          clearNextRoundTimeout();
          setState((s) => ({
            ...s,
            currentTurn: nextStartingPlayer ?? s.currentTurn,
            roundNumber,
            board: emptyBoard,
            hand: [],
            // No handsSize map on this event, and HandDealt is per-player
            // only — assume the standard equal deal size for every seat
            // until this player's own hand arrives via HandDealt.
            handCounts: Object.fromEntries(
              s.playerOrder.map((id) => [id, STARTING_HAND_SIZE]),
            ),
            roundOver: null,
            latestRoundOverPoints: null,
            nextRoundRequested: false,
          }));
          break;
        }
        case GameEvents.RoundOver: {
          const data = message.data;
          setState((s) => {
            const pointsThisRound: Record<string, number> = {};
            for (const team of Object.keys(data.gameScores)) {
              pointsThisRound[team] =
                data.gameScores[team] - (s.gameScores[team] ?? 0);
            }
            const historyEntry: RoundHistoryEntry = {
              roundNumber: data.roundNumber,
              roundID: data.roundID,
              roundResult: data.roundResult,
              gameScoresAfter: data.gameScores,
              pointsThisRound,
            };
            return {
              ...s,
              roundOver: data,
              gameScores: data.gameScores,
              gameStatus: data.gameState,
              teamWinner: data.teamWinner || null,
              latestRoundOverPoints: pointsThisRound,
              roundHistory: [...s.roundHistory, historyEntry],
            };
          });
          break;
        }
        case GameEvents.NextRoundResponse: {
          clearNextRoundTimeout();
          setState((s) => ({ ...s, nextRoundRequested: false }));
          break;
        }
        case LobbyEvents.PlayerConnected: {
          const { userID } = message.data;
          setState((s) => ({
            ...s,
            playerConnectivity: { ...s.playerConnectivity, [userID]: true },
          }));
          break;
        }
        case LobbyEvents.PlayerDisconnected: {
          const { userID } = message.data;
          setState((s) => ({
            ...s,
            playerConnectivity: { ...s.playerConnectivity, [userID]: false },
          }));
          break;
        }
        case LobbyEvents.PlayerJoined: {
          // playerCount is the joining player's slot: JoinLobby assigns
          // slot = len(players)+1 and publishes after adding, so the two
          // always match (see services/lobby-service/internal/service/lobby.go).
          const { userID: joinedID, displayName, playerCount } = message.data;
          if (joinedID === userID) break;
          setState((s) =>
            s.joinedPlayers.some((p) => p.id === joinedID)
              ? s
              : {
                  ...s,
                  joinedPlayers: [
                    ...s.joinedPlayers,
                    {
                      id: joinedID,
                      name: displayName,
                      slot: playerCount,
                      isConnected: false,
                    },
                  ],
                },
          );
          break;
        }
      }
    };

    return () => {
      wsRef.current = null;
      clearNextRoundTimeout();
      if (ws.readyState === WebSocket.OPEN) ws.close();
    };
  }, [lobbyID, wsToken, userID]);

  const playTile = (tile: Tile, side: Side) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setState((s) => ({ ...s, error: "not connected" }));
      return;
    }
    ws.send(
      JSON.stringify({ type: GameEvents.PlayTileCmd, data: { tile, side } }),
    );
  };

  const pass = () => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setState((s) => ({ ...s, error: "not connected" }));
      return;
    }
    ws.send(JSON.stringify({ type: GameEvents.PassTurnCmd, data: {} }));
  };

  const nextRound = () => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setState((s) => ({ ...s, error: "not connected" }));
      return;
    }
    setState((s) => ({ ...s, nextRoundRequested: true }));
    ws.send(JSON.stringify({ type: GameEvents.NextRoundCmd, data: {} }));

    // The gateway silently drops a rejected NextRoundCmd (no error is ever
    // sent back), so fall back to clearing the pending state ourselves if
    // neither RoundStarted nor NextRoundResponse arrives in time.
    clearNextRoundTimeout();
    nextRoundTimeoutRef.current = setTimeout(() => {
      setState((s) => ({
        ...s,
        nextRoundRequested: false,
        error: "couldn't start next round — try again",
      }));
    }, NEXT_ROUND_TIMEOUT_MS);
  };

  return { ...state, playTile, pass, nextRound };
}
