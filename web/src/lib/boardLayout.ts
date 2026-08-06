// Derived rendering geometry for Board.tsx. Never feeds back into board.ts —
// applyMove/legalSides mirror the backend's replay logic (see board.ts's
// header comment) and their tiles/leftEnd/rightEnd contract must stay
// untouched.
//
// Layout rules — a spiral, not a grid:
//  1. The tile that opens the round is placed dead center and never moves
//     again — every later tile is laid out *around* it, not the whole chain
//     re-centered. The board grows outward from the center tile's right
//     edge (toward the chain's right/open end) and from its left edge
//     (toward the chain's left/open end), independently.
//  2. Each side has a current travel direction, starting horizontal — right
//     for the right-growing side, left for the left-growing side — and
//     keeps placing tiles edge-to-edge in that direction for as long as
//     they fit inside the board's fixed square canvas.
//  3. The instant the *next* tile would cross the canvas border, the
//     side's travel direction rotates 90° counter-clockwise (right → up →
//     left → down → right) and the chain continues from exactly the point
//     it stopped — no gap, no re-centering, no dedicated "corner" tile.
//     Both sides use the same counter-clockwise rule; they start 180°
//     apart, so they trace mirror arcs of the same spiral and never cross.
//  4. Doubles are always laid crosswise (perpendicular to the *current*
//     travel direction), same as real play, whether or not they happen to
//     land on a turn. That's the only tile that ever isn't aligned with the
//     current direction — a turn itself doesn't need a special pivot tile,
//     since every plain tile already renders aligned with whatever the
//     (possibly just-rotated) direction is.
// Because the board only ever grows by exactly one tile at a time (see
// useGameConnection: every state update is a single applyMove), positions
// are computed incrementally and cached in a ref keyed by tile identity
// (`left-right` is unique — a standard set has no duplicate tiles, and a
// placed tile's orientation never changes after the fact) so that already-
// placed tiles never move on screen as new ones are added.
import { useRef } from "react";
import { BoardState } from "./board";
import { Tile, tileToString } from "./types";

// Pixel geometry for the "sm" DominoTile (see DominoTile.tsx's SIZE_CLASSES /
// ROTATED_SIZE_CLASSES). Kept in sync manually — there are only the three
// fixed sizes and Board.tsx always renders "sm" here.
const TILE_LONG = 80; // footprint along the direction of travel, laid straight
const TILE_SHORT = 40; // footprint along the direction of travel, laid crosswise
const GAP = 4; // matches Tailwind's gap-1
const SLOT_SIZE = 40; // EndSlot button diameter
// Breathing room kept between the outermost tile edge and the felt border —
// place()/seed() are bounded to `canvasHalf - MARGIN`, while the rendered
// canvas itself stays the full `canvasHalf * 2`, so tiles visibly stop short
// of the rim instead of touching or crossing it.
const MARGIN = 40;

export interface LaidOutTile {
  key: string;
  tile: Tile;
  rotation: number;
  left: number;
  top: number;
}

export interface EndSlotPixel {
  left: number;
  top: number;
}

export interface BoardLayout {
  width: number;
  height: number;
  tiles: LaidOutTile[];
  leftEndSlot: EndSlotPixel;
  rightEndSlot: EndSlotPixel;
}

type Dir = "R" | "U" | "L" | "D";
const VECTORS: Record<Dir, { x: number; y: number }> = {
  R: { x: 1, y: 0 },
  U: { x: 0, y: -1 },
  L: { x: -1, y: 0 },
  D: { x: 0, y: 1 },
};
// 90° counter-clockwise.
const ROTATE_CCW: Record<Dir, Dir> = { R: "U", U: "L", L: "D", D: "R" };

// Screen-clockwise angle (E=0, S=90, W=180, N=270) that a non-double tile
// must be rotated to so its "open" pip half — the one that continues the
// chain further outward — faces the direction the arm is currently
// traveling. DominoTile always draws tile.left/tile.right as a west/east
// pair before any rotation is applied, and board.ts's applyMove defines
// which of those two values is "open" differently per side (see below), so
// the same travel direction needs a different rotation on each side.
const FORWARD_ANGLE: Record<Dir, number> = { R: 0, D: 90, L: 180, U: 270 };

// `side` mirrors board.ts's applyMove: the bottom cursor grows the chain's
// right/open end (where a played tile's *right* value is the new open end
// and *left* is what connects backward), the top cursor grows the left/open
// end (mirrored — *left* is open, *right* connects backward). Doubles show
// the same pip both halves, so their rotation only needs to satisfy the
// footprint's aspect ratio, not this forward/backward pip placement.
function computeRotation(isDouble: boolean, width: number, side: "top" | "bottom", dir: Dir): number {
  if (isDouble) return width === TILE_SHORT ? 90 : 0;
  const forward = FORWARD_ANGLE[dir];
  return side === "bottom" ? forward : (forward + 180) % 360;
}

interface Cursor {
  x: number;
  y: number;
  dir: Dir;
}

interface Box {
  left: number;
  top: number;
  width: number;
  height: number;
}

interface Placed extends Box {
  rotation: number;
}

interface LayoutState {
  placed: Map<string, Placed>;
  order: string[]; // board.tiles keys, in board order, as of the last call
  bottom: Cursor; // grows toward the chain's right/open-end end
  top: Cursor; // grows toward the chain's left/open-start end
}

function emptyState(): LayoutState {
  return {
    placed: new Map(),
    order: [],
    bottom: { x: 0, y: 0, dir: "R" },
    top: { x: 0, y: 0, dir: "L" },
  };
}

function box(x: number, y: number, dir: Dir, footprint: number, perp: number): Box {
  switch (dir) {
    case "R":
      return { left: x, top: y - perp / 2, width: footprint, height: perp };
    case "L":
      return { left: x - footprint, top: y - perp / 2, width: footprint, height: perp };
    case "D":
      return { left: x - perp / 2, top: y, width: perp, height: footprint };
    case "U":
      return { left: x - perp / 2, top: y - footprint, width: perp, height: footprint };
  }
}

// Advances `cursor` by one tile. If the tile's full footprint — not just its
// leading edge along the travel direction — would cross the canvas border,
// the direction rotates 90° counter-clockwise (possibly more than once, for
// a canvas too small to fit even one tile in any direction) and the tile is
// placed in the new direction instead, from the exact point the cursor
// stopped. Checking the whole box (not just the forward endpoint) matters
// for the tile that turns: see the refX/refY comment below — the pivot shift
// can push a turning tile's perpendicular extent past the border even when
// the straight-line forward check alone would've allowed it.
function place(
  cursor: Cursor,
  tile: Tile,
  half: number,
  side: "top" | "bottom"
): { placed: Placed; next: Cursor } {
  const isDouble = tile.left === tile.right;
  const footprint = isDouble ? TILE_SHORT : TILE_LONG;
  const perp = isDouble ? TILE_LONG : TILE_SHORT;

  const priorDir = cursor.dir;
  let dir = cursor.dir;
  let vec = VECTORS[dir];
  let refX = cursor.x;
  let refY = cursor.y;
  let b: Box;
  for (let attempt = 0; attempt < 4; attempt++) {
    vec = VECTORS[dir];

    // box() centers a tile on (x, y) across its perpendicular axis. That's
    // right for a tile continuing the same direction — it shares the run's
    // established centerline — but wrong for the tile that just turned: its
    // predecessor's footprint occupies the half of that centerline behind
    // the pivot (the direction we arrived from), so centering on the pivot
    // would have the new tile's back half overlap it. Shifting the
    // reference point half a tile-width further in the *prior* direction
    // makes the new tile's trailing edge land exactly on the pivot instead,
    // GAP away from what it's turning off of, regardless of either tile's
    // size.
    refX = cursor.x;
    refY = cursor.y;
    if (dir !== priorDir) {
      const priorVec = VECTORS[priorDir];
      refX += priorVec.x * (perp / 2);
      refY += priorVec.y * (perp / 2);
    }

    b = box(refX, refY, dir, footprint, perp);
    const overflows = b.left < -half || b.top < -half || b.left + b.width > half || b.top + b.height > half;
    if (!overflows || attempt === 3) break;
    dir = ROTATE_CCW[dir];
  }

  const step = footprint + GAP;
  const next: Cursor = { x: refX + vec.x * step, y: refY + vec.y * step, dir };

  const rotation = computeRotation(isDouble, b!.width, side, dir);
  return { placed: { ...b!, rotation }, next };
}

// (Re)seeds `state` treating board.tiles[0] as the opening, centered tile
// and every subsequent tile as a rightward/downward-spiral extension. Used
// both for the very first tile of a round and as a fallback if the board's
// tiles ever change in a way that isn't recognized as simple growth (e.g. a
// resync).
function seed(state: LayoutState, board: BoardState, half: number, keys: string[]) {
  const first = board.tiles[0];
  const isDouble = first.left === first.right;
  const footprint = isDouble ? TILE_SHORT : TILE_LONG;
  const perp = isDouble ? TILE_LONG : TILE_SHORT;
  const b = box(-footprint / 2, 0, "R", footprint, perp);
  state.placed.set(keys[0], { ...b, rotation: computeRotation(isDouble, b.width, "bottom", "R") });
  state.bottom = { x: footprint / 2 + GAP, y: 0, dir: "R" };
  state.top = { x: -footprint / 2 - GAP, y: 0, dir: "L" };

  for (let i = 1; i < board.tiles.length; i++) {
    const { placed, next } = place(state.bottom, board.tiles[i], half, "bottom");
    state.placed.set(keys[i], placed);
    state.bottom = next;
  }
  state.order = keys;
}

export function useBoardLayout(board: BoardState, opts: { canvas: number }): BoardLayout | null {
  // `canvasHalf` sizes the rendered surface (buildLayout); `half` is the
  // smaller boundary tiles are actually placed within, so the outermost
  // tile stops MARGIN short of the felt edge instead of touching/crossing
  // it (see place()'s box-overflow check).
  const canvasHalf = Math.max(TILE_LONG, opts.canvas) / 2;
  const half = canvasHalf - MARGIN;
  const stateRef = useRef<LayoutState | null>(null);

  if (board.tiles.length === 0) {
    stateRef.current = null;
    return null;
  }

  const state = stateRef.current ?? (stateRef.current = emptyState());
  const keys = board.tiles.map(tileToString);

  const grewOnBottom =
    state.order.length > 0 &&
    keys.length >= state.order.length &&
    keys.slice(0, state.order.length).join("|") === state.order.join("|");
  const grewOnTop =
    !grewOnBottom &&
    state.order.length > 0 &&
    keys.length >= state.order.length &&
    keys.slice(keys.length - state.order.length).join("|") === state.order.join("|");

  if (state.order.length === 0) {
    seed(state, board, half, keys);
  } else if (grewOnBottom) {
    for (let i = state.order.length; i < keys.length; i++) {
      const { placed, next } = place(state.bottom, board.tiles[i], half, "bottom");
      state.placed.set(keys[i], placed);
      state.bottom = next;
    }
    state.order = keys;
  } else if (grewOnTop) {
    const grewCount = keys.length - state.order.length;
    // Prepended tiles appear at indices [0, grewCount) in board order, but
    // the one *closest* to the previously-known chain (index grewCount - 1)
    // is the one that actually attaches to it, so it must be placed first.
    for (let i = grewCount - 1; i >= 0; i--) {
      const { placed, next } = place(state.top, board.tiles[i], half, "top");
      state.placed.set(keys[i], placed);
      state.top = next;
    }
    state.order = keys;
  } else if (keys.join("|") !== state.order.join("|")) {
    // Board changed in a way that isn't simple growth (new round, resync) —
    // rebuild from scratch rather than show stale positions.
    stateRef.current = emptyState();
    seed(stateRef.current, board, half, keys);
  }

  return buildLayout(board, stateRef.current, canvasHalf);
}

// The canvas is a fixed `2*half` square regardless of how many tiles are
// placed — every tile/slot is offset by the constant `half` (not by the
// bounding box of tiles placed so far) so the rendered surface never grows
// or shrinks as the round progresses; it matches the fixed half-plane the
// spiral placement math (`place()`) already confines every tile to.
function buildLayout(board: BoardState, state: LayoutState, half: number): BoardLayout {
  const tiles: LaidOutTile[] = board.tiles.map((tile) => {
    const key = tileToString(tile);
    const p = state.placed.get(key)!;
    return { key, tile, rotation: p.rotation, left: p.left + half, top: p.top + half };
  });

  const slotBox = (cursor: Cursor): EndSlotPixel => {
    const b = box(cursor.x, cursor.y, cursor.dir, SLOT_SIZE, SLOT_SIZE);
    return { left: b.left + half, top: b.top + half };
  };

  return {
    width: half * 2,
    height: half * 2,
    tiles,
    leftEndSlot: slotBox(state.top),
    rightEndSlot: slotBox(state.bottom),
  };
}
