import { cn } from "@/lib/utils";
import { PlayerModel } from "@/lib/types";

interface PlayerSeatProps {
  player: PlayerModel | undefined;
  handCount: number;
  isCurrentTurn: boolean;
  isSelf: boolean;
}

const MAX_TILE_BACKS = 7;

export function PlayerSeat({ player, handCount, isCurrentTurn, isSelf }: PlayerSeatProps) {
  const name = player?.name ?? "…";

  return (
    <div
      className={cn(
        "flex flex-col items-center gap-1 rounded-lg border-2 bg-card p-2 text-center transition-colors",
        isCurrentTurn ? "border-primary shadow-md" : "border-border"
      )}
    >
      <div
        className={cn(
          "flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold",
          isCurrentTurn ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
        )}
      >
        {player ? player.name.slice(0, 2).toUpperCase() : "?"}
      </div>
      <p className="max-w-20 truncate text-xs font-medium">
        {name}
        {isSelf && <span className="text-muted-foreground"> (you)</span>}
      </p>
      <TileBacks count={handCount} />
    </div>
  );
}

function TileBacks({ count }: { count: number }) {
  const shown = Math.min(count, MAX_TILE_BACKS);
  return (
    <div className="flex items-center gap-1">
      <div className="flex">
        {Array.from({ length: shown }).map((_, i) => (
          <div key={i} className="-ml-2 h-6 w-3 shrink-0 rounded-sm border bg-muted first:ml-0" />
        ))}
      </div>
      {count > MAX_TILE_BACKS && <span className="text-xs text-muted-foreground">+{count - MAX_TILE_BACKS}</span>}
      {count === 0 && <span className="text-xs text-muted-foreground">0</span>}
    </div>
  );
}
