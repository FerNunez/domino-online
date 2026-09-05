import { DominoTile } from "./DominoTile";
import { PlayerChip } from "./PlayerChip";
import { cn } from "@/lib/utils";
import { PlayerModel, Tile, tileToString } from "@/lib/types";
import { SeatEdge } from "@/lib/seatLayout";

interface PlayerSeatProps {
  player: PlayerModel | undefined;
  handCount: number;
  // This player's revealed tiles once the round is over (RoundOverData.playerHands).
  // Undefined mid-round — the rack stays face-down via handCount as before.
  revealedTiles?: Tile[];
  isCurrentTurn: boolean;
  isSelf: boolean;
  isYourTeam?: boolean;
  edge: SeatEdge;
}

// A full starting hand is 7 — showing every one of them (rather than a
// truncated "+N") is the point of this rack: at a glance it tells you how
// close an opponent is to going out.
const MAX_TILE_BACKS = 7;

export function PlayerSeat({ player, handCount, revealedTiles, isCurrentTurn, isSelf, isYourTeam, edge }: PlayerSeatProps) {
  const name = player?.name ?? "…";
  const vertical = edge === "left" || edge === "right";

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
      {revealedTiles ? <TileFaces tiles={revealedTiles} vertical={vertical} /> : <TileBacks count={handCount} vertical={vertical} />}
    </div>
  );
}

// A fanned rack of face-down tiles, one per tile actually left in the
// opponent's hand — same idea as Hand's face-up tiles, just without pips
// (we never learn an opponent's tiles until they're played). Count alone
// (as digits) is harder to read at a glance than a rack shrinking tile by
// tile, which is the whole point of showing it this way.
//
// Left/right seats spread the rack vertically (like a real tile rack held
// side-on) instead of horizontally. The gap between tiles is computed so a
// full 7-tile hand fills RACK_SPAN_PX with tight overlap, and the rack
// stays visually "full" (rather than shrinking to a tiny cluster) as tiles
// are played, capped at one tile's thickness so a near-empty hand doesn't
// spread into an oddly huge gap.
const RACK_SPAN_PX = 112;
const TILE_SPAN_PX = 16;
const OVERLAP_MIN_PX = -10;

function TileBacks({ count, vertical = false }: { count: number; vertical?: boolean }) {
  const shown = Math.min(count, MAX_TILE_BACKS);
  // Step between consecutive tile origins (not the margin itself — see
  // below) so a full 7-tile hand's last tile ends exactly at RACK_SPAN_PX.
  const step =
    shown > 1
      ? Math.min(TILE_SPAN_PX, Math.max(OVERLAP_MIN_PX, (RACK_SPAN_PX - TILE_SPAN_PX) / (shown - 1)))
      : 0;
  // marginTop/Left sits *on top of* the tile's own footprint, so it must be
  // the step minus that footprint (negative once step < TILE_SPAN_PX, i.e.
  // genuine overlap) — applying `step` directly here would double-count
  // each tile's footprint and overflow the rack's declared span.
  const margin = step - TILE_SPAN_PX;

  return (
    <div className={cn("flex items-center", vertical ? "h-28 w-8 flex-col justify-center" : "h-8")}>
      {Array.from({ length: shown }).map((_, i) => (
        <div
          key={i}
          style={i === 0 ? undefined : vertical ? { marginTop: margin } : { marginLeft: margin }}
          className={cn(
            "flex shrink-0 items-center justify-center rounded-sm border-2 border-border bg-muted shadow-sm",
            vertical ? "h-4 w-8" : "h-8 w-4"
          )}
        >
          <div className="h-1.5 w-1.5 rotate-45 rounded-[1px] bg-border" />
        </div>
      ))}
      {count === 0 && <span className="text-xs text-muted-foreground">0 tiles</span>}
    </div>
  );
}

// The round-over reveal of this seat's actual tiles. Real pips (via
// DominoTile) need real pixels — legible dots start at DominoTile's smallest
// built-in size, well past the tile-back rack's tiny 16px footprint — so
// this wraps onto multiple rows (or stacks into a column, for left/right
// seats) instead of trying to fan/overlap like TileBacks does. The seat box
// grows to fit at the moment of reveal, which is fine: it only happens once
// the round has already ended. Tiles stay in their natural flat orientation
// (same as Hand and RoundHandsReveal), no rotation needed now that nothing
// has to fan tightly.
function TileFaces({ tiles, vertical = false }: { tiles: Tile[]; vertical?: boolean }) {
  if (tiles.length === 0) {
    return <p className="text-xs font-medium">domino!</p>;
  }

  return (
    <div className={cn("flex items-center justify-center gap-1", vertical ? "flex-col" : "max-w-[168px] flex-wrap")}>
      {tiles.map((tile, i) => (
        <DominoTile key={`${tileToString(tile)}-${i}`} tile={tile} size="sm" />
      ))}
    </div>
  );
}
