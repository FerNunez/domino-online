import { Button } from "@/components/ui/button";
import { RoundOverData } from "@/lib/types";

interface RoundSummaryProps {
  roundOver: RoundOverData;
  yourTeamID?: string;
  canStartNextRound: boolean;
  nextRoundRequested: boolean;
  pointsThisRound: Record<string, number> | null;
  onNextRound: () => void;
}

const REASON_VERB: Record<string, string> = {
  REASON_DOMINO: "domino",
  REASON_BLOCKED: "block",
};

// Always-visible status text in the board's top-center HUD slot (see
// Board/GameScreen) — never a popup. The board itself, and the revealed
// hands in RoundHandsReveal, stay fully visible and reachable underneath.
// Rendered as plain text (no card/pill) since it's pure readout; only the
// actual Next round control keeps real button styling, so it stays visually
// distinct as the one clickable thing in this row.
export function RoundSummary({
  roundOver,
  yourTeamID,
  canStartNextRound,
  nextRoundRequested,
  pointsThisRound,
  onNextRound,
}: RoundSummaryProps) {
  const { roundResult, gameState, teamWinner } = roundOver;
  const matchFinished = gameState === "GAME_STATUS_FINISHED";

  const youWonMatch = !!teamWinner && teamWinner === yourTeamID;
  const youWonRound = !!roundResult.winnerTeamID && roundResult.winnerTeamID === yourTeamID;
  const verb = REASON_VERB[roundResult.reason ?? ""] ?? "points";
  const points = roundResult.winnerTeamID ? (pointsThisRound?.[roundResult.winnerTeamID] ?? 0) : 0;

  const statusText = matchFinished
    ? youWonMatch
      ? "You won the match! 🏆"
      : "The other team won the match"
    : !roundResult.winnerTeamID
      ? "Round tied"
      : youWonRound
        ? `You won by ${verb} — +${points} points`
        : `The other team won by ${verb} — +${points} points`;

  return (
    <div className="flex items-center gap-2 text-xs font-medium [text-shadow:0_1px_2px_rgb(0_0_0/0.25)]">
      <span>{statusText}</span>
      {!matchFinished &&
        (canStartNextRound ? (
          <Button size="sm" className="h-6 rounded-full px-2 text-xs" disabled={nextRoundRequested} onClick={onNextRound}>
            {nextRoundRequested ? "Starting…" : "Next round"}
          </Button>
        ) : (
          <span className="text-muted-foreground">waiting on host…</span>
        ))}
    </div>
  );
}
