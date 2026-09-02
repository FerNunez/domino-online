import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface TurnIndicatorProps {
  currentTurn: string | null;
  // Display name of the current-turn player, when known — falls back to the
  // raw id if the roster lookup misses (shouldn't normally happen).
  currentTurnName?: string;
  // False only when we have a live/snapshot signal that this player is
  // disconnected — see GameScreen's fallback chain.
  currentTurnConnected: boolean;
  userID: string | undefined;
  hasLegalMove: boolean;
  onPass: () => void;
}

export function TurnIndicator({ currentTurn, currentTurnName, currentTurnConnected, userID, hasLegalMove, onPass }: TurnIndicatorProps) {
  const isYourTurn = !!currentTurn && currentTurn === userID;
  const waitingOnDisconnected = !isYourTurn && !!currentTurn && !currentTurnConnected;
  const displayName = currentTurnName ?? currentTurn;

  return (
    <div className={cn("flex items-center justify-between gap-4 rounded-lg border bg-card p-3", waitingOnDisconnected && "border-amber-500 bg-amber-500/10")}>
      <p className="text-sm font-medium">
        {isYourTurn
          ? "Your turn"
          : currentTurn
            ? waitingOnDisconnected
              ? `Waiting for ${displayName} to reconnect…`
              : `Waiting on ${displayName}`
            : "Waiting for game to start…"}
      </p>
      <Button size="sm" variant="secondary" disabled={!isYourTurn || hasLegalMove} onClick={onPass}>
        Pass
      </Button>
    </div>
  );
}
