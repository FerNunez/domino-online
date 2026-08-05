# Board tile placement

How `useBoardLayout` (`boardLayout.ts`) decides where each tile on the board
renders. This is purely a rendering concern — it never feeds back into
`board.ts`, which mirrors the backend's chain logic (`tiles`/`leftEnd`/
`rightEnd`) and must stay untouched.

## The model: a spiral, not a grid

The chain is drawn as two straight-line paths growing outward from the
opening tile — one toward the chain's left/open end, one toward its right —
each path turning 90° when it runs out of room, so the whole board traces a
rectangular spiral around the board's edge.

1. **The opening tile is fixed.** It's placed dead center the moment it's
   played and never moves again. Every later tile is laid out *around* it —
   the whole chain is never re-centered when it grows.
2. **Growth is independent per side.** The left-open-end path and the
   right-open-end path each track their own position and travel direction;
   neither affects the other except that they share the same canvas (see
   [Collision safety](#collision-safety-between-the-two-sides) below).
3. **Each side starts horizontal** — the right-growing side travels right,
   the left-growing side travels left — and keeps placing tiles edge-to-edge
   in that direction for as long as they fit inside the board's fixed square
   canvas.
4. **Hitting the edge turns the path, not the tile.** The instant the *next*
   tile would cross the canvas border, that side's travel direction rotates
   **90° counter-clockwise** (right → up → left → down → right) and the
   chain continues from exactly the point it stopped. No gap, no
   re-centering, and no dedicated "corner" tile — the turn is a property of
   the path, not a special tile.
5. **Doubles are always crosswise.** A double is laid perpendicular to
   whichever direction the path is currently traveling, whether or not it
   happens to land on a turn — this is the one case where a tile isn't
   aligned with the current travel direction. Every other tile, including
   the one right after a turn, is simply aligned with the (possibly
   just-rotated) direction.

## Constraints this produces

- **Tiles never move once placed.** The board only ever grows by exactly one
  tile per update (see `useGameConnection` — every state change is a single
  `applyMove`), so positions are computed incrementally and cached by tile
  identity (`left-right`, unique — a standard set has no duplicate tiles,
  and a placed tile's orientation never changes after the fact). A tile's
  final rendered position is decided the moment it's placed and is never
  recomputed afterward.
- **Every touching pair is exactly `GAP` (4px) apart** — the same spacing as
  any two in-line tiles — whether that's two tiles in a straight run or a
  tile meeting the one that just turned off in a new direction. There's no
  such thing as a "looser" corner gap.
- **A turn never overlaps the tile it turned away from.** The tile
  immediately after a turn doesn't get centered on the turn point — its
  trailing edge (the side facing back toward where the path came from) is
  placed exactly on that point, so it only ever extends into new space. This
  holds regardless of the sizes of the two tiles involved.

## Collision safety between the two sides

The left- and right-growing paths share one canvas and turn the same way
(counter-clockwise), so they trace mirror arcs of the same spiral rather
than converging — but for a long enough game, a side that's turned a full
loop can come back around close to where the *other* side started. The
canvas is sized (800px, 10 tile-widths) so that a full 26–28 tile game
can't do this: verified by replaying 2,000 randomly generated valid games
through the layout and checking for overlaps.

If tiles are ever swapped for a different size, or the canvas made
significantly smaller, that safety margin should be re-checked — it isn't
enforced structurally, just currently wide enough for a full double-six
game.
