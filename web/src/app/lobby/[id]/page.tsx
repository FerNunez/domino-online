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

  const { lobby: lobbySnapshot, error: lobbyError } = useLobbySnapshot(lobbyID);
  const conn = useGameConnection(lobbyID, wsToken, userID);

  // The snapshot is a one-shot fetch (see useLobbySnapshot); players who join
  // afterwards arrive live via lobby.player_joined and are merged in here.
  const lobby =
    lobbySnapshot && conn.joinedPlayers.length > 0
      ? {
          ...lobbySnapshot,
          players: [
            ...lobbySnapshot.players,
            ...conn.joinedPlayers.filter((jp) => !lobbySnapshot.players.some((p) => p.id === jp.id)),
          ],
        }
      : lobbySnapshot;

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
        {lobby ? (
          <GameScreen
            userID={userID}
            lobby={lobby}
            playerOrder={conn.playerOrder}
            handCounts={conn.handCounts}
            board={conn.board}
            hand={conn.hand}
            currentTurn={conn.currentTurn}
            roundOver={conn.roundOver}
            gameScores={conn.gameScores}
            roundNumber={conn.roundNumber}
            latestRoundOverPoints={conn.latestRoundOverPoints}
            nextRoundRequested={conn.nextRoundRequested}
            isHost={isHost}
            playTile={conn.playTile}
            pass={conn.pass}
            nextRound={conn.nextRound}
          />
        ) : (
          <p className="text-sm text-muted-foreground">Loading players…</p>
        )}
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
