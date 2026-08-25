"use client";

import { useState } from "react";
import { RoundActionsList } from "./RoundActionsList";
import { getRoundActions } from "@/lib/api";
import { GameRoundSummary, HistoryAction, HistoryHand } from "@/lib/types";
import { cn } from "@/lib/utils";

interface GameRoundRowProps {
  round: GameRoundSummary;
  yourTeamID?: string;
  playerName: (id: string) => string;
}

const REASON_COPY: Record<string, string> = {
  REASON_DOMINO: "by dominoing out",
  REASON_BLOCKED: "board blocked, lowest pips wins",
};

// One round in a game's history, expandable in place to lazy-load and
// replay its move-by-move actions (GET /rounds/{id}/actions) — most rounds
// in a game are never expanded, so fetching upfront for every row would
// waste requests.
export function GameRoundRow({ round, yourTeamID, playerName }: GameRoundRowProps) {
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [actions, setActions] = useState<HistoryAction[] | null>(null);
  const [hands, setHands] = useState<HistoryHand[]>([]);

  const toggle = async () => {
    const next = !expanded;
    setExpanded(next);
    if (next && actions === null && !loading) {
      setLoading(true);
      setError(null);
      try {
        const data = await getRoundActions(round.roundId);
        setActions(data.actions);
        setHands(data.hands);
      } catch (err) {
        setError(err instanceof Error ? err.message : "failed to load round actions");
      } finally {
        setLoading(false);
      }
    }
  };

  const winnerLabel = !round.winnerTeamId
    ? "Tied"
    : yourTeamID
      ? round.winnerTeamId === yourTeamID
        ? "Your team won"
        : "Opponents won"
      : `${round.winnerTeamId} won`;

  return (
    <li className="rounded-lg border bg-card">
      <button
        type="button"
        onClick={toggle}
        className="flex w-full items-center justify-between gap-4 p-3 text-left"
      >
        <div>
          <p className="text-sm font-medium">
            Round {round.roundNumber} — {winnerLabel}
          </p>
          <p className="text-xs text-muted-foreground">
            {REASON_COPY[round.reason ?? ""] ?? round.reason ?? "unknown reason"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <ul className="flex gap-3 text-xs text-muted-foreground">
            {Object.entries(round.scores).map(([teamID, score]) => (
              <li key={teamID}>
                {teamID}: {score}
              </li>
            ))}
          </ul>
          <span className={cn("text-xs transition-transform", expanded && "rotate-180")}>▾</span>
        </div>
      </button>

      {expanded && (
        <div className="border-t p-3">
          {loading && <p className="text-sm text-muted-foreground">Loading moves…</p>}
          {error && <p className="text-sm text-destructive">{error}</p>}
          {actions && <RoundActionsList actions={actions} hands={hands} playerName={playerName} />}
        </div>
      )}
    </li>
  );
}
