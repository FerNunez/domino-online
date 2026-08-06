import { cn } from "@/lib/utils";
import { Tile } from "@/lib/types";

// Classic six-face pip layout on a 3x3 grid (indices 0-8, row-major).
const PIP_LAYOUTS: Record<number, number[]> = {
  0: [],
  1: [4],
  2: [0, 8],
  3: [0, 4, 8],
  4: [0, 2, 6, 8],
  5: [0, 2, 4, 6, 8],
  6: [0, 2, 3, 5, 6, 8],
};

function PipHalf({ value }: { value: number }) {
  const filled = new Set(PIP_LAYOUTS[value] ?? []);
  return (
    <div className="grid h-full w-full grid-cols-3 grid-rows-3 gap-0.5 p-1">
      {Array.from({ length: 9 }).map((_, i) => (
        <div key={i} className="flex items-center justify-center">
          {filled.has(i) && <div className="h-1.5 w-1.5 rounded-full bg-foreground sm:h-2 sm:w-2" />}
        </div>
      ))}
    </div>
  );
}

interface DominoTileProps {
  tile: Tile;
  size?: "sm" | "md" | "lg";
  // Clockwise degrees (0/90/180/270) to rotate the whole tile as a rigid
  // piece — pip halves stay left/right internally, they just read as
  // top/bottom (and, at 180, swap left-right) once rotated. Used both for
  // doubles laid across the chain and for tiles on a vertical spiral run,
  // where the exact angle (not just 90-or-not) is what keeps the matching
  // pip facing the neighbor it actually connects to.
  rotation?: number;
  selected?: boolean;
  disabled?: boolean;
  onClick?: () => void;
  className?: string;
}

const SIZE_CLASSES: Record<NonNullable<DominoTileProps["size"]>, string> = {
  sm: "h-10 w-20",
  md: "h-14 w-28",
  lg: "h-16 w-32",
};

// h/w swapped so a rotated tile's bounding box reflects its actual footprint
// in the layout (a rotated "sm" tile is 40px wide, not 80px).
const ROTATED_SIZE_CLASSES: Record<NonNullable<DominoTileProps["size"]>, string> = {
  sm: "h-20 w-10",
  md: "h-28 w-14",
  lg: "h-32 w-16",
};

export function DominoTile({ tile, size = "md", rotation = 0, selected, disabled, onClick, className }: DominoTileProps) {
  const interactive = !!onClick && !disabled;
  // 90/270 swap the footprint's bounding box; 0/180 keep it as drawn.
  const aspectSwapped = rotation === 90 || rotation === 270;
  return (
    <div
      className={cn("flex shrink-0 items-center justify-center", aspectSwapped ? ROTATED_SIZE_CLASSES[size] : undefined)}
    >
      <div
        role={interactive ? "button" : undefined}
        tabIndex={interactive ? 0 : undefined}
        onClick={interactive ? onClick : undefined}
        onKeyDown={interactive ? (e) => (e.key === "Enter" || e.key === " ") && onClick?.() : undefined}
        style={rotation ? { transform: `rotate(${rotation}deg)` } : undefined}
        className={cn(
          "flex shrink-0 overflow-hidden rounded-md border-2 bg-card shadow-sm transition-all duration-150",
          SIZE_CLASSES[size],
          interactive && "cursor-pointer hover:-translate-y-1 hover:shadow-md",
          selected && "-translate-y-2 border-primary shadow-md",
          disabled && "opacity-40",
          !selected && !disabled && "border-border",
          className
        )}
      >
        <PipHalf value={tile.left} />
        <div className="w-px shrink-0 bg-border" />
        <PipHalf value={tile.right} />
      </div>
    </div>
  );
}
