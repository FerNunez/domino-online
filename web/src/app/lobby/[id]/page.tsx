"use client";

import { use, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { LobbyRoster } from "@/components/LobbyRoster";
import { GameScreen } from "@/components/GameScreen";
import { useLobbySnapshot } from "@/hooks/useLobbySnapshot";
import { useGameConnection } from "@/hooks/useGameConnection";
import { getStoredToken, getStoredWsToken, startLobby } from "@/lib/api";
import { decodeJwtUserId } from "@/lib/jwt";

// The WebSocket connection is opened here (not in a separate /game/[id] route)
// so `game.game_started` / `game.hand_dealt` are never missed while waiting —
// see the plan's note on lifting the connection instead of racing a reconnect
// on navigation. /game/[id] exists as a fallback for direct links (see that page).
export default function LobbyPage({ params }: { params: Promise<{ id: string }> }) {
  const { id: lobbyID } = use(params);
  const [userID, setUserID] = useState<string | undefined>(undefined);
  const [wsToken, setWsToken] = useState<string | undefined>(undefined);
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = getStoredToken();
    if (token) setUserID(decodeJwtUserId(token) ?? undefined);
    setWsToken(getStoredWsToken(lobbyID) ?? undefined);
  }, [lobbyID]);

  const { lobby, error: lobbyError } = useLobbySnapshot(lobbyID);
  const conn = useGameConnection(lobbyID, wsToken, userID);

  const isHost = !!lobby && !!userID && lobby.hostId === userID;
  const isFull = !!lobby && lobby.players.length === lobby.maxPlayers;

  const handleStart = async () => {
    setStarting(true);
    setError(null);
    try {
      await startLobby(lobbyID);
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to start game");
      setStarting(false);
    }
  };

  if (conn.gameStarted) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center">
        <GameScreen
          userID={userID}
          board={conn.board}
          hand={conn.hand}
          currentTurn={conn.currentTurn}
          roundResult={conn.roundResult}
          playTile={conn.playTile}
          pass={conn.pass}
        />
      </main>
    );
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-6 p-6">
      <h1 className="text-2xl font-bold">Lobby {lobbyID}</h1>
      <p className="text-sm text-muted-foreground">Share this lobby ID with 3 friends to fill the table.</p>

      {lobby ? (
        <LobbyRoster lobby={lobby} playerConnectivity={conn.playerConnectivity} />
      ) : (
        <p className="text-sm text-muted-foreground">Loading lobby…</p>
      )}

      {isHost && (
        <Button onClick={handleStart} disabled={!isFull || starting}>
          {starting ? "Starting…" : isFull ? "Start game" : "Waiting for players…"}
        </Button>
      )}

      {(error || lobbyError || conn.error) && (
        <p className="text-sm text-destructive">{error ?? lobbyError ?? conn.error}</p>
      )}
    </main>
  );
}
