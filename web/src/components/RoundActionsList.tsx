import { DominoTile } from "./DominoTile";
import { HistoryAction, HistoryHand } from "@/lib/types";

interface RoundActionsListProps {
  actions: HistoryAction[];
  hands?: HistoryHand[];
  playerName: (id: string) => string;
}

function ActionRow({ action, playerName }: { action: HistoryAction; playerName: (id: string) => string }) {
  const name = playerName(action.playerId);
  if (action.actionType === "pass" || !action.tile) {
    return (
      <li className="flex items-center gap-2 py-1 text-sm text-muted-foreground">
        <span className="w-6 text-right text-xs">{action.actionNumber}.</span>
        {name} passed
      </li>
    );
  }
  return (
    <li className="flex items-center gap-2 py-1 text-sm">
      <span className="w-6 text-right text-xs text-muted-foreground">{action.actionNumber}.</span>
      <DominoTile tile={action.tile} size="sm" />
      <span>
        {name} played on the {action.side ?? "?"}
      </span>
    </li>
  );
}

// Shared by GameRoundRow (post-game history browsing) and RoundActionsPanel
// (the live game screen, right after a round ends) — both render the same
// move-by-move log and starting hands from GET /rounds/{id}/actions.
export function RoundActionsList({ actions, hands, playerName }: RoundActionsListProps) {
  return (
    <div className="flex flex-col gap-3">
      {hands && hands.length > 0 && (
        <div className="flex flex-col gap-2">
          <p className="text-xs font-medium text-muted-foreground">Starting hands</p>
          <ul className="flex flex-col gap-1">
            {hands.map((hand) => (
              <li key={hand.playerId} className="flex items-center gap-2">
                <span className="w-24 shrink-0 truncate text-xs text-muted-foreground">{playerName(hand.playerId)}</span>
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
        <ul>
          {actions.map((a) => (
            <ActionRow key={a.actionNumber} action={a} playerName={playerName} />
          ))}
        </ul>
      )}
    </div>
  );
}
