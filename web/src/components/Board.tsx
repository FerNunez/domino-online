import { ReactNode } from "react";
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
  // HUD overlays pinned to the four corners/top-center of the board itself
  // (score, round counter, turn status, pass button) — kept as generic slots
  // so Board doesn't need to know what's inside them.
  scoreboard?: ReactNode;
  roundBadge?: ReactNode;
  turnStatus?: ReactNode;
  passButton?: ReactNode;
}

export function Board({ board, canPlayLeft, canPlayRight, onDropEnd, scoreboard, roundBadge, turnStatus, passButton }: BoardProps) {
  const layout = useBoardLayout(board, { canvas: CANVAS });

  const hud = (
    <>
      {scoreboard && <div className="absolute left-3 top-3">{scoreboard}</div>}
      {roundBadge && <div className="absolute right-3 top-3">{roundBadge}</div>}
      {turnStatus && <div className="absolute left-1/2 top-3 -translate-x-1/2">{turnStatus}</div>}
      {passButton && <div className="absolute bottom-3 right-3">{passButton}</div>}
    </>
  );

  return (
    <div className="flex w-full items-center justify-center overflow-auto rounded-lg p-6">
      <div
        className="relative shrink-0 rounded-2xl border-4 border-amber-900/20 bg-emerald-800/10 shadow-inner dark:border-amber-100/10 dark:bg-emerald-900/20"
        style={{ width: layout.width, height: layout.height }}
      >
        {board.tiles.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center">
            <p className="text-sm text-muted-foreground">Board is empty — play any tile to start</p>
          </div>
        )}
        {layout.tiles.map((lt) => (
          <div key={lt.key} className="absolute" style={{ left: lt.left, top: lt.top }}>
            <DominoTile tile={lt.tile} size="sm" rotation={lt.rotation} />
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
        {hud}
      </div>
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
