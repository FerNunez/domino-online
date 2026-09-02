import { DominoTile } from "./DominoTile";
import { PlayerChip } from "./PlayerChip";
import { HistoryAction, HistoryHand } from "@/lib/types";

interface RoundActionsListProps {
  actions: HistoryAction[];
  hands?: HistoryHand[];
  playerName: (id: string) => string;
}

// "left"/"right" describe which end of the chain the tile extended, same
// vocabulary as the board itself — an arrow reads faster than the sentence
// it replaces ("played on the left") once there are two dozen of these.
const SIDE_ARROW: Record<string, string> = { left: "←", right: "→" };

function ActionRow({ action, playerName }: { action: HistoryAction; playerName: (id: string) => string }) {
  const name = playerName(action.playerId);
  return (
    <li className="grid grid-cols-[1.5rem_1.25rem_1fr_1.25rem] items-center gap-2 py-1 text-sm" title={name}>
      <span className="text-right text-xs text-muted-foreground">{action.actionNumber}.</span>
      <PlayerChip name={name} size="sm" />
      {action.actionType === "pass" || !action.tile ? (
        <span className="col-span-2 text-muted-foreground italic">passed</span>
      ) : (
        <>
          <DominoTile tile={action.tile} size="sm" />
          <span className="text-muted-foreground" aria-label={`played on the ${action.side ?? "?"}`}>
            {SIDE_ARROW[action.side ?? ""] ?? action.side ?? "?"}
          </span>
        </>
      )}
    </li>
  );
}

// Shared by GameRoundRow (post-game history browsing) and RoundActionsPanel
// (the live game screen, right after a round ends) — both render the same
// move-by-move log and starting hands from GET /rounds/{id}/actions. The
// actions list caps its own height and scrolls internally (lichess/chess.com-
// style) rather than growing the page — a blocked round can run 25+ actions.
export function RoundActionsList({ actions, hands, playerName }: RoundActionsListProps) {
  return (
    <div className="flex flex-col gap-3">
      {hands && hands.length > 0 && (
        <div className="flex flex-col gap-2">
          <p className="text-xs font-medium text-muted-foreground">Starting hands</p>
          <ul className="flex flex-col gap-1">
            {hands.map((hand) => (
              <li key={hand.playerId} className="flex items-center gap-2">
                <PlayerChip name={playerName(hand.playerId)} size="sm" />
                <div className="flex flex-wrap gap-1">
                  {hand.tiles.map((tile, i) => (
                    <DominoTile key={i} tile={tile} size="sm" />
                  ))}
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {actions.length === 0 ? (
        <p className="text-sm text-muted-foreground">No moves recorded for this round.</p>
      ) : (
        <ul className="max-h-56 overflow-y-auto pr-1">
          {actions.map((a) => (
            <ActionRow key={a.actionNumber} action={a} playerName={playerName} />
          ))}
        </ul>
      )}
    </div>
  );
}
