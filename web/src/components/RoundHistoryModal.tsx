"use client";

import { useState } from "react";
import { NotebookPen } from "lucide-react";
import { RoundActionsList } from "./RoundActionsList";
import { Button } from "./ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./ui/dialog";
import { ApiError, getRoundActions } from "@/lib/api";
import { HistoryAction, HistoryHand, PlayerModel, RoundHistoryEntry } from "@/lib/types";
import { cn } from "@/lib/utils";

interface RoundHistoryModalProps {
  roundHistory: RoundHistoryEntry[];
  players: PlayerModel[];
  yourTeamID?: string;
}

const REASON_COPY: Record<string, string> = {
  REASON_DOMINO: "by dominoing out",
  REASON_BLOCKED: "board blocked, lowest pips wins",
};

// Sits in the board's corner action cluster next to the pass/knock button
// (see GameScreen) — clicking it never affects the live game, only opens a
// read-only look back at rounds already played this game (client-accumulated
// in useGameConnection's roundHistory; see that type for why it's
// session-local, not durable).
export function RoundHistoryModal({ roundHistory, players, yourTeamID }: RoundHistoryModalProps) {
  const playerName = (id: string) => players.find((p) => p.id === id)?.name ?? `${id.slice(0, 8)}…`;
  const roundsNewestFirst = [...roundHistory].reverse();

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          type="button"
          size="icon"
          variant="secondary"
          className="rounded-full shadow-sm"
          title="Round history"
          aria-label="Round history"
          disabled={roundHistory.length === 0}
        >
          <NotebookPen />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Round history</DialogTitle>
        </DialogHeader>
        {roundsNewestFirst.length === 0 ? (
          <p className="text-sm text-muted-foreground">No rounds finished yet.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {roundsNewestFirst.map((entry) => (
              <RoundHistoryRow key={entry.roundID} entry={entry} playerName={playerName} yourTeamID={yourTeamID} />
            ))}
          </ul>
        )}
      </DialogContent>
    </Dialog>
  );
}

function RoundHistoryRow({
  entry,
  playerName,
  yourTeamID,
}: {
  entry: RoundHistoryEntry;
  playerName: (id: string) => string;
  yourTeamID?: string;
}) {
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
        const data = await getRoundActions(entry.roundID);
        setActions(data.actions);
        setHands(data.hands);
      } catch (err) {
        const notReadyYet = err instanceof ApiError && err.status === 404;
        setError(notReadyYet ? "still saving this round…" : err instanceof Error ? err.message : "failed to load");
      } finally {
        setLoading(false);
      }
    }
  };

  const winnerTeamID = entry.roundResult.winnerTeamID;
  const winnerLabel = !winnerTeamID
    ? "Tied"
    : yourTeamID
      ? winnerTeamID === yourTeamID
        ? "Your team won"
        : "Opponents won"
      : `${winnerTeamID} won`;

  return (
    <li className="rounded-lg border bg-card">
      <button type="button" onClick={toggle} className="flex w-full items-center justify-between gap-4 p-3 text-left">
        <div>
          <p className="text-sm font-medium">
            Round {entry.roundNumber} — {winnerLabel}
          </p>
          <p className="text-xs text-muted-foreground">
            {REASON_COPY[entry.roundResult.reason ?? ""] ?? entry.roundResult.reason ?? "unknown reason"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <ul className="flex gap-3 text-xs text-muted-foreground">
            {Object.entries(entry.pointsThisRound).map(([teamID, points]) => (
              <li key={teamID}>
                {teamID}: +{points}
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
