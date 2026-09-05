import { useState } from "react";
import { Trophy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { PlayerModel, RoundOverData, slotToTeamID } from "@/lib/types";
import { cn } from "@/lib/utils";

interface RoundSummaryProps {
  roundOver: RoundOverData;
  yourTeamID?: string;
  players: PlayerModel[];
  canStartNextRound: boolean;
  nextRoundRequested: boolean;
  pointsThisRound: Record<string, number> | null;
  onNextRound: () => void;
}

const REASON_COPY: Record<string, string> = {
  REASON_DOMINO: "by dominoing out",
  REASON_BLOCKED: "board blocked, lowest pips wins",
};

// Sits in the board's top-center HUD slot (replacing TurnStatus once a round
// ends) rather than covering the board — the board and the tiles already
// played stay visible and reachable behind it. Tap the status text to see
// the full score breakdown in a dialog; the next-round control stays on the
// bar itself so the host doesn't need an extra click to advance the game.
export function RoundSummary({
  roundOver,
  yourTeamID,
  players,
  canStartNextRound,
  nextRoundRequested,
  pointsThisRound,
  onNextRound,
}: RoundSummaryProps) {
  const [detailsOpen, setDetailsOpen] = useState(false);
  const { roundResult, gameScores, gameState, teamWinner } = roundOver;
  const matchFinished = gameState === "GAME_STATUS_FINISHED";

  const teamLabel = (teamID: string) => {
    const names = players.filter((p) => slotToTeamID(p.slot) === teamID).map((p) => p.name);
    return names.length > 0 ? names.join(" & ") : teamID;
  };

  const youWonMatch = !!teamWinner && teamWinner === yourTeamID;
  const youWonRound = !!roundResult.winnerTeamID && roundResult.winnerTeamID === yourTeamID;

  const statusText = matchFinished
    ? youWonMatch
      ? "Match over — you won! 🏆"
      : "Match over — you lost"
    : roundResult.winnerTeamID
      ? youWonRound
        ? "Round over — you won! 🎉"
        : "Round over — you lost"
      : "Round over — tied";

  const dialogTitle = matchFinished
    ? youWonMatch
      ? "Your team won the match! 🏆"
      : "The other team won the match"
    : roundResult.winnerTeamID
      ? youWonRound
        ? "You won this round! 🎉"
        : "The other team won this round"
      : "Round tied";

  return (
    <>
      <div className="flex items-center gap-2 rounded-full bg-card/90 px-3 py-1.5 text-xs font-medium shadow-sm">
        <button type="button" onClick={() => setDetailsOpen(true)} className="hover:underline">
          {statusText}
        </button>
        {!matchFinished &&
          (canStartNextRound ? (
            <Button size="sm" className="h-6 px-2 text-xs" disabled={nextRoundRequested} onClick={onNextRound}>
              {nextRoundRequested ? "Starting…" : "Next round"}
            </Button>
          ) : (
            <span className="text-muted-foreground">waiting on host…</span>
          ))}
      </div>

      <Dialog open={detailsOpen} onOpenChange={setDetailsOpen}>
        <DialogContent className="text-center sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="text-center">{dialogTitle}</DialogTitle>
          </DialogHeader>
          {!matchFinished && (
            <p className="text-sm text-muted-foreground">
              {REASON_COPY[roundResult.reason ?? ""] ?? roundResult.reason ?? "unknown reason"}
            </p>
          )}
          <ul className="flex flex-col gap-1 text-sm">
            {matchFinished
              ? Object.entries(gameScores).map(([teamID, score]) => (
                  <TeamRow key={teamID} label={teamLabel(teamID)} isWinner={teamID === teamWinner}>
                    {score}
                  </TeamRow>
                ))
              : pointsThisRound &&
                Object.entries(pointsThisRound).map(([teamID, points]) => (
                  <TeamRow key={teamID} label={teamLabel(teamID)} isWinner={teamID === roundResult.winnerTeamID}>
                    +{points} (total {gameScores[teamID] ?? 0})
                  </TeamRow>
                ))}
          </ul>
        </DialogContent>
      </Dialog>
    </>
  );
}

function TeamRow({ label, isWinner, children }: { label: string; isWinner: boolean; children: React.ReactNode }) {
  return (
    <li className={cn("flex items-center justify-center gap-1.5", isWinner ? "font-semibold" : "text-muted-foreground")}>
      {isWinner && <Trophy className="h-3.5 w-3.5 text-amber-500" />}
      <span>
        {label}: {children}
      </span>
    </li>
  );
}
