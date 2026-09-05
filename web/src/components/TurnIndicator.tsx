import { cn } from "@/lib/utils";

interface TurnStatusProps {
  currentTurn: string | null;
  // Display name of the current-turn player, when known — falls back to the
  // raw id if the roster lookup misses (shouldn't normally happen).
  currentTurnName?: string;
  // False only when we have a live/snapshot signal that this player is
  // disconnected — see GameScreen's fallback chain.
  currentTurnConnected: boolean;
  userID: string | undefined;
}

// Text-only turn status, overlaid on the board (see Board's hud slots) as
// plain text rather than a card/pill — it's pure readout, never interactive,
// and a filled background would read as a clickable chip. The drop-shadow
// keeps it legible over the felt; the disconnected warning is called out by
// color alone rather than a border/background box. The Pass button is
// rendered separately by GameScreen so it can sit in its own corner rather
// than sharing a row with this text.
export function TurnStatus({ currentTurn, currentTurnName, currentTurnConnected, userID }: TurnStatusProps) {
  const isYourTurn = !!currentTurn && currentTurn === userID;
  const waitingOnDisconnected = !isYourTurn && !!currentTurn && !currentTurnConnected;
  const displayName = currentTurnName ?? currentTurn;

  return (
    <p
      className={cn(
        "text-xs font-medium [text-shadow:0_1px_2px_rgb(0_0_0/0.25)]",
        waitingOnDisconnected ? "text-amber-600 dark:text-amber-400" : "text-foreground"
      )}
    >
      {isYourTurn
        ? "Your turn"
        : currentTurn
          ? waitingOnDisconnected
            ? `Waiting for ${displayName} to reconnect…`
            : `Waiting on ${displayName}`
          : "Waiting for game to start…"}
    </p>
  );
}
