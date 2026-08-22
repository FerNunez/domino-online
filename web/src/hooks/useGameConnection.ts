import { useEffect, useRef, useState } from "react";
import { WEBSOCKET_URL } from "@/lib/constants";
import {
  GameEvents,
  isValidServerWsMessage,
  LobbyEvents,
  ServerWsMessage,
} from "@/lib/contracts";
import { applyMove, BoardState, emptyBoard } from "@/lib/board";
import { PlayerModel, RoundResult, Side, Tile } from "@/lib/types";

interface GameConnectionState {
  connected: boolean;
  gameStarted: boolean;
  playerOrder: string[];
  currentTurn: string | null;
  hand: Tile[];
  board: BoardState;
  roundResult: RoundResult | null;
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
  roundResult: null,
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
          const { playerOrder, currentTurn, handsSize } = message.data;
          setState((s) => ({
            ...s,
            gameStarted: true,
            playerOrder,
            currentTurn,
            board: emptyBoard,
            roundResult: null,
            handCounts: handsSize ?? {},
          }));
          break;
        }
        case GameEvents.HandDealt: {
          if (message.data.playerID !== userID) break;
          setState((s) => ({ ...s, hand: message.data.playerTiles }));
          break;
        }
        case GameEvents.PlayerMoveMade: {
          const {
            userID: actorID,
            tile,
            side,
            nextTurn: nextTurn,
            roundResult: roundResult,
          } = message.data;
          setState((s) => ({
            ...s,
            board: applyMove(s.board, tile, side as Side),
            currentTurn: nextTurn,
            roundResult: roundResult ?? s.roundResult,
            handCounts: {
              ...s.handCounts,
              [actorID]: Math.max(0, (s.handCounts[actorID] ?? 0) - 1),
            },
          }));
          break;
        }
        case GameEvents.PlayerPassed: {
          const { next_turn, round_result: roundResult } = message.data;
          setState((s) => ({
            ...s,
            currentTurn: next_turn,
            roundResult: roundResult ?? s.roundResult,
          }));
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

  return { ...state, playTile, pass };
}
