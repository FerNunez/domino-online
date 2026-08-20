import { Button } from "@/components/ui/button";

interface TurnIndicatorProps {
  currentTurn: string | null;
  userID: string | undefined;
  hasLegalMove: boolean;
  onPass: () => void;
}

export function TurnIndicator({ currentTurn, userID, hasLegalMove, onPass }: TurnIndicatorProps) {
  const isYourTurn = !!currentTurn && currentTurn === userID;

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
