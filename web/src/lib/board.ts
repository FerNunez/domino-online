// Mirrors GameModel.PlayTile's board placement logic in
// services/game-service/internal/domain/game.go:167-239. The `game.player_move_made`
// broadcast only carries the incremental {tile, side} the player submitted (not
// the resulting board), so every client replays moves locally to derive the
// same oriented board the server holds. There is no GET-game endpoint to
// re-sync from, so this replay must stay in lockstep with the domain logic above.
import { Side, Tile } from "./types";

export interface BoardState {
  tiles: Tile[];
  leftEnd: number;
  rightEnd: number;
}

export const emptyBoard: BoardState = { tiles: [], leftEnd: -1, rightEnd: -1 };

function flip(tile: Tile): Tile {
  return { left: tile.right, right: tile.left };
}

function hasPip(tile: Tile, pip: number): boolean {
  return tile.left === pip || tile.right === pip;
}

// Mirrors ValidMoves in services/game-service/internal/domain/game.go:137-155.
export function legalSides(board: BoardState, tile: Tile): Side[] {
  if (board.tiles.length === 0) return ["left"];
  const sides: Side[] = [];
  if (hasPip(tile, board.leftEnd)) sides.push("left");
  if (hasPip(tile, board.rightEnd)) sides.push("right");
  return sides;
}

export function applyMove(board: BoardState, tile: Tile, side: Side): BoardState {
  if (board.tiles.length === 0) {
    return { tiles: [tile], leftEnd: tile.left, rightEnd: tile.right };
  }

  if (side === "left") {
    const oriented = tile.right === board.leftEnd ? tile : flip(tile);
    return {
      tiles: [oriented, ...board.tiles],
      leftEnd: oriented.left,
      rightEnd: board.rightEnd,
    };
  }

  const oriented = tile.left === board.rightEnd ? tile : flip(tile);
  return {
    tiles: [...board.tiles, oriented],
    leftEnd: board.leftEnd,
    rightEnd: oriented.right,
  };
}
