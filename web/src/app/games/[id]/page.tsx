"use client";

import Link from "next/link";
import { use, useEffect, useMemo, useState } from "react";
import { GameRoundRow } from "@/components/GameRoundRow";
import { getGameHistory, getLobby, getPlayerGames, getStoredToken } from "@/lib/api";
import { decodeJwtUserId } from "@/lib/jwt";
import { GameRoundSummary, GameSummary, LobbyModel, slotToTeamID } from "@/lib/types";

// Per-game round breakdown (GET /games/{id}/history), with per-round move
// replay on demand via GameRoundRow. Reached from /games; the ?lobbyID query
// param (carried over from that list, where GameSummary.lobbyId is already
// known) lets this page fetch the lobby roster to resolve player names/teams
// — without it, rounds still render, just with raw player IDs.
export default function GameHistoryPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ lobbyID?: string }>;
}) {
  const { id: gameID } = use(params);
  const { lobbyID } = use(searchParams);

  const [rounds, setRounds] = useState<GameRoundSummary[] | null>(null);
  const [summary, setSummary] = useState<GameSummary | null>(null);
  const [lobby, setLobby] = useState<LobbyModel | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [userID, setUserID] = useState<string | undefined>(undefined);

  useEffect(() => {
    const token = getStoredToken();
    if (token) setUserID(decodeJwtUserId(token) ?? undefined);
  }, []);

  useEffect(() => {
    let cancelled = false;
    getGameHistory(gameID)
      .then((data) => {
        if (!cancelled) setRounds(data);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "failed to load game history");
      });
    return () => {
      cancelled = true;
    };
  }, [gameID]);

  useEffect(() => {
    if (!lobbyID) return;
    let cancelled = false;
    getLobby(lobbyID)
      .then((data) => {
        if (!cancelled) setLobby(data);
      })
      .catch(() => {
        // best-effort — round rows fall back to raw player IDs
      });
    return () => {
      cancelled = true;
    };
  }, [lobbyID]);

  // GetGameHistoryResponse only carries rounds, not game-level fields (final
  // score, winner, created date) — pull those from the recent-games list
  // instead of adding a dedicated backend endpoint just for this header.
  useEffect(() => {
    let cancelled = false;
    getPlayerGames()
      .then((games) => {
        if (!cancelled) setSummary(games.find((g) => g.gameId === gameID) ?? null);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [gameID]);

  const yourTeamID = useMemo(() => {
    if (!lobby || !userID) return undefined;
    const slot = lobby.players.find((p) => p.id === userID)?.slot;
    return slot !== undefined ? slotToTeamID(slot) : undefined;
  }, [lobby, userID]);

  const playerName = useMemo(() => {
    const byID = new Map((lobby?.players ?? []).map((p) => [p.id, p.name]));
    return (id: string) => byID.get(id) ?? `${id.slice(0, 8)}…`;
  }, [lobby]);

  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Game history</h1>
        <Link href="/games" className="text-sm text-muted-foreground underline-offset-4 hover:underline">
          Back to My Games
        </Link>
      </div>

      {summary && (
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm font-medium">
            {summary.gameState === "GAME_STATUS_FINISHED"
              ? !summary.teamWinner
                ? "Tied"
                : summary.teamWinner === yourTeamID
                  ? "Your team won"
                  : `${summary.teamWinner} won`
              : "In progress"}
          </p>
          <ul className="flex gap-3 text-sm text-muted-foreground">
            {Object.entries(summary.finalScores).map(([teamID, score]) => (
              <li key={teamID}>
                {teamID === yourTeamID ? "Your Team" : teamID}: {score}
              </li>
            ))}
          </ul>
        </div>
      )}

      {error && <p className="text-sm text-destructive">{error}</p>}
      {!error && rounds === null && <p className="text-sm text-muted-foreground">Loading rounds…</p>}
      {rounds && rounds.length === 0 && (
        <p className="text-sm text-muted-foreground">No rounds recorded for this game.</p>
      )}

      <ul className="flex flex-col gap-2">
        {rounds?.map((round) => (
          <GameRoundRow key={round.roundId} round={round} yourTeamID={yourTeamID} playerName={playerName} />
        ))}
      </ul>
    </main>
  );
}
