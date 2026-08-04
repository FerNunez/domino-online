import { DominoTile } from "./DominoTile";
import { BoardState } from "@/lib/board";
import { Side, tileToString } from "@/lib/types";

interface BoardProps {
  board: BoardState;
  canPlayLeft: boolean;
  canPlayRight: boolean;
  onDropEnd: (side: Side) => void;
}

export function Board({ board, canPlayLeft, canPlayRight, onDropEnd }: BoardProps) {
  return (
    <div className="flex min-h-32 w-full items-center justify-center gap-1 overflow-x-auto rounded-lg border bg-muted/30 p-6">
      {board.tiles.length === 0 ? (
        <p className="text-sm text-muted-foreground">Board is empty — play any tile to start</p>
      ) : (
        <>
          <EndSlot active={canPlayLeft} onClick={() => onDropEnd("left")} label={String(board.leftEnd)} />
          {board.tiles.map((tile, i) => (
            <DominoTile key={`${tileToString(tile)}-${i}`} tile={tile} size="sm" />
          ))}
          <EndSlot active={canPlayRight} onClick={() => onDropEnd("right")} label={String(board.rightEnd)} />
        </>
      )}
    </div>
  );
}

function EndSlot({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  if (!active) return null;
  return (
    <button
      onClick={onClick}
      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border-2 border-dashed border-primary text-xs font-medium text-primary transition-colors hover:bg-primary/10"
      title={`Play here (matches ${label})`}
    >
      {label}
    </button>
  );
}
