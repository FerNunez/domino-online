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

export function DominoTile({ tile, size = "md", selected, disabled, onClick, className }: DominoTileProps) {
  const interactive = !!onClick && !disabled;
  return (
    <div
      role={interactive ? "button" : undefined}
      tabIndex={interactive ? 0 : undefined}
      onClick={interactive ? onClick : undefined}
      onKeyDown={interactive ? (e) => (e.key === "Enter" || e.key === " ") && onClick?.() : undefined}
      className={cn(
        "flex overflow-hidden rounded-md border-2 bg-card shadow-sm transition-all duration-150",
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
  );
}
