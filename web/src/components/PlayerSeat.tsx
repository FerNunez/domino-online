import { PlayerChip } from "./PlayerChip";
import { cn } from "@/lib/utils";
import { PlayerModel } from "@/lib/types";

interface PlayerSeatProps {
  player: PlayerModel | undefined;
  handCount: number;
  isCurrentTurn: boolean;
  isSelf: boolean;
  isYourTeam?: boolean;
}

// A full starting hand is 7 — showing every one of them (rather than a
// truncated "+N") is the point of this rack: at a glance it tells you how
// close an opponent is to going out.
const MAX_TILE_BACKS = 7;

export function PlayerSeat({ player, handCount, isCurrentTurn, isSelf, isYourTeam }: PlayerSeatProps) {
  const name = player?.name ?? "…";

  return (
    <div
      className={cn(
        "flex flex-col items-center gap-1 rounded-lg border-2 border-l-4 bg-card p-2 text-center transition-colors",
        isCurrentTurn ? "border-primary shadow-md" : "border-border",
        isYourTeam === true && "border-l-blue-500",
        isYourTeam === false && "border-l-amber-500"
      )}
    >
      <PlayerChip name={player?.name ?? "?"} active={isCurrentTurn} />
      <p className="max-w-20 truncate text-xs font-medium">
        {name}
        {isSelf && <span className="text-muted-foreground"> (you)</span>}
      </p>
      <TileBacks count={handCount} />
    </div>
  );
}

// A fanned rack of face-down tiles, one per tile actually left in the
// opponent's hand — same idea as Hand's face-up tiles, just without pips
// (we never learn an opponent's tiles until they're played). Count alone
// (as digits) is harder to read at a glance than a rack shrinking tile by
// tile, which is the whole point of showing it this way.
function TileBacks({ count }: { count: number }) {
  const shown = Math.min(count, MAX_TILE_BACKS);
  return (
    <div className="flex h-8 items-center">
      {Array.from({ length: shown }).map((_, i) => (
        <div
          key={i}
          className="-ml-2.5 flex h-8 w-4 shrink-0 items-center justify-center rounded-sm border-2 border-border bg-muted shadow-sm first:ml-0"
        >
          <div className="h-1.5 w-1.5 rotate-45 rounded-[1px] bg-border" />
        </div>
      ))}
      {count === 0 && <span className="text-xs text-muted-foreground">0 tiles</span>}
    </div>
  );
}
