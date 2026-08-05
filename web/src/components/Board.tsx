import { DominoTile } from "./DominoTile";
import { BoardState } from "@/lib/board";
import { useBoardLayout } from "@/lib/boardLayout";
import { Side } from "@/lib/types";
import { cn } from "@/lib/utils";

// Static approximation tuned to the "sm" tile size (80px) — how far a
// straight run travels before it hits the canvas edge and spirals. A
// ResizeObserver-driven value would adapt to the actual container size, but
// isn't needed for a first pass. 800px (10 tiles) rather than a tighter fit
// — the two independent spirals (one per open end) share this canvas and
// need enough room that a fully-wrapped side can't loop back around into
// the other side's starting area for a full 28-tile game.
const CANVAS = 10 * 80;

interface BoardProps {
  board: BoardState;
  canPlayLeft: boolean;
  canPlayRight: boolean;
  onDropEnd: (side: Side) => void;
}

export function Board({ board, canPlayLeft, canPlayRight, onDropEnd }: BoardProps) {
  const layout = useBoardLayout(board, { canvas: CANVAS });

  return (
    <div className="flex min-h-32 w-full items-center justify-center overflow-auto rounded-lg border bg-muted/30 p-6">
      {!layout ? (
        <p className="text-sm text-muted-foreground">Board is empty — play any tile to start</p>
      ) : (
        <div className="relative" style={{ width: layout.width, height: layout.height }}>
          {layout.tiles.map((lt) => (
            <div key={lt.key} className="absolute" style={{ left: lt.left, top: lt.top }}>
              <DominoTile tile={lt.tile} size="sm" rotate={lt.rotate} />
            </div>
          ))}
          {canPlayLeft && (
            <EndSlot
              left={layout.leftEndSlot.left}
              top={layout.leftEndSlot.top}
              onClick={() => onDropEnd("left")}
              label={String(board.leftEnd)}
            />
          )}
          {canPlayRight && (
            <EndSlot
              left={layout.rightEndSlot.left}
              top={layout.rightEndSlot.top}
              onClick={() => onDropEnd("right")}
              label={String(board.rightEnd)}
            />
          )}
        </div>
      )}
    </div>
  );
}

function EndSlot({ left, top, onClick, label }: { left: number; top: number; onClick: () => void; label: string }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "absolute flex h-10 w-10 items-center justify-center rounded-full border-2 border-dashed border-primary text-xs font-medium text-primary transition-colors hover:bg-primary/10"
      )}
      style={{ left, top }}
      title={`Play here (matches ${label})`}
    >
      {label}
    </button>
  );
}
