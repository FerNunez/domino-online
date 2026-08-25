"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ensureGuestToken, getPlayerGames } from "@/lib/api";
import { GameSummary } from "@/lib/types";

function formatDate(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString();
}

// Lists the caller's most recent finished/in-progress games, durably
// persisted by history-service — distinct from useGameConnection's
// in-memory roundHistory, which only covers the game currently live on a
// websocket connection.
export default function GamesPage() {
  const [games, setGames] = useState<GameSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        await ensureGuestToken();
        const data = await getPlayerGames();
        if (!cancelled) setGames(data);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "failed to load games");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">My Games</h1>
        <Link href="/" className="text-sm text-muted-foreground underline-offset-4 hover:underline">
          Back home
        </Link>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}
      {!error && games === null && <p className="text-sm text-muted-foreground">Loading games…</p>}
      {games && games.length === 0 && (
        <p className="text-sm text-muted-foreground">No games played yet.</p>
      )}

      <ul className="flex flex-col gap-2">
        {games?.map((game) => (
          <li key={game.gameId}>
            <Link
              href={`/games/${game.gameId}?lobbyID=${game.lobbyId}`}
              className="flex items-center justify-between gap-4 rounded-lg border bg-card p-4 hover:bg-accent"
            >
              <div>
                <p className="text-sm font-medium">
                  {game.gameState === "GAME_STATUS_FINISHED"
                    ? game.teamWinner
                      ? `${game.teamWinner} won`
                      : "Tied"
                    : "In progress"}
                </p>
                <p className="text-xs text-muted-foreground">{formatDate(game.createdAt)}</p>
              </div>
              <ul className="flex gap-3 text-sm text-muted-foreground">
                {Object.entries(game.finalScores).map(([teamID, score]) => (
                  <li key={teamID}>
                    {teamID}: {score}
                  </li>
                ))}
              </ul>
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
