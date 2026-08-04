import { Button } from "@/components/ui/button";
import { RoundResult } from "@/lib/types";

interface TurnIndicatorProps {
  currentTurn: string | null;
  userID: string | undefined;
  hasLegalMove: boolean;
  roundResult: RoundResult | null;
  onPass: () => void;
}

export function TurnIndicator({ currentTurn, userID, hasLegalMove, roundResult, onPass }: TurnIndicatorProps) {
  const isYourTurn = !!currentTurn && currentTurn === userID;

  if (roundResult) {
    const youWon = roundResult.winner_id === userID;
    return (
      <div className="flex flex-col items-center gap-2 rounded-lg border bg-card p-6 text-center">
        <h2 className="text-xl font-semibold">
          {roundResult.winner_id ? (youWon ? "You won! 🎉" : `${roundResult.winner_id} won the round`) : "Round tied"}
        </h2>
        <p className="text-sm text-muted-foreground">reason: {roundResult.reason ?? "unknown"}</p>
        {roundResult.scores && (
          <ul className="text-sm">
            {Object.entries(roundResult.scores).map(([id, score]) => (
              <li key={id}>
                {id}: {score}
              </li>
            ))}
          </ul>
        )}
      </div>
    );
  }

  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border bg-card p-3">
      <p className="text-sm font-medium">
        {isYourTurn ? "Your turn" : currentTurn ? `Waiting on ${currentTurn}` : "Waiting for game to start…"}
      </p>
      <Button size="sm" variant="secondary" disabled={!isYourTurn || hasLegalMove} onClick={onPass}>
        Pass
      </Button>
    </div>
  );
}
