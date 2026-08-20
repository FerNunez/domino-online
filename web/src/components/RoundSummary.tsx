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

const REASON_COPY: Record<string, string> = {
  REASON_DOMINO: "by dominoing out",
  REASON_BLOCKED: "board blocked, lowest pips wins",
};

export function RoundSummary({
  roundOver,
  yourTeamID,
  canStartNextRound,
  nextRoundRequested,
  pointsThisRound,
  onNextRound,
}: RoundSummaryProps) {
  const { roundResult, gameScores, gameState, teamWinner } = roundOver;
  const matchFinished = gameState === "GAME_STATUS_FINISHED";

  if (matchFinished) {
    const youWon = !!teamWinner && teamWinner === yourTeamID;
    return (
      <div className="flex flex-col items-center gap-2 rounded-lg border bg-card p-6 text-center">
        <h2 className="text-xl font-semibold">
          {youWon ? "Your team won the match! 🏆" : "The other team won the match"}
        </h2>
        <ul className="text-sm">
          {Object.entries(gameScores).map(([teamID, score]) => (
            <li key={teamID}>
              {teamID === yourTeamID ? "Your Team" : "Opponents"}: {score}
            </li>
          ))}
        </ul>
      </div>
    );
  }

  const youWonRound = !!roundResult.winnerTeamID && roundResult.winnerTeamID === yourTeamID;

  return (
    <div className="flex flex-col items-center gap-2 rounded-lg border bg-card p-6 text-center">
      <h2 className="text-xl font-semibold">
        {roundResult.winnerTeamID
          ? youWonRound
            ? "You won this round! 🎉"
            : "The other team won this round"
          : "Round tied"}
      </h2>
      <p className="text-sm text-muted-foreground">
        {REASON_COPY[roundResult.reason ?? ""] ?? roundResult.reason ?? "unknown reason"}
      </p>
      {pointsThisRound && (
        <ul className="text-sm">
          {Object.entries(pointsThisRound).map(([teamID, points]) => (
            <li key={teamID}>
              {teamID === yourTeamID ? "Your Team" : "Opponents"}: +{points} (total {gameScores[teamID] ?? 0})
            </li>
          ))}
        </ul>
      )}
      {canStartNextRound ? (
        <Button size="sm" disabled={nextRoundRequested} onClick={onNextRound}>
          {nextRoundRequested ? "Starting…" : "Next Round"}
        </Button>
      ) : (
        <p className="text-sm text-muted-foreground">Waiting for the host to start the next round…</p>
      )}
    </div>
  );
}
