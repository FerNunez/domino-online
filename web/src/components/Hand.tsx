import { DominoTile } from "./DominoTile";
import { Tile } from "@/lib/types";
import { tileToString } from "@/lib/types";

interface HandProps {
  tiles: Tile[];
  selectedTile?: Tile | null;
  isYourTurn: boolean;
  onSelect: (tile: Tile) => void;
}

export function Hand({ tiles, selectedTile, isYourTurn, onSelect }: HandProps) {
  return (
    <div className="flex flex-wrap items-end justify-center gap-2 p-4">
      {tiles.map((tile, i) => (
        <DominoTile
          key={`${tileToString(tile)}-${i}`}
          tile={tile}
          size="md"
          disabled={!isYourTurn}
          selected={!!selectedTile && tileToString(selectedTile) === tileToString(tile)}
          onClick={() => onSelect(tile)}
        />
      ))}
      {tiles.length === 0 && <p className="text-sm text-muted-foreground">No tiles left</p>}
    </div>
  );
}
