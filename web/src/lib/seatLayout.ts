// Pure seat-assignment for the in-game player ring (PlayersPanel.tsx). The
// server's playerOrder is already the authoritative clockwise turn sequence
// (playerOrder[i+1] plays after playerOrder[i]), so rendering it rotated so
// the local player is last/bottom, then walking left -> top -> right, is all
// that's needed to make the ring read clockwise starting from the bottom.
export type SeatEdge = "bottom" | "left" | "top" | "right";

export interface SeatAssignment {
  id: string;
  edge: SeatEdge;
}

export function assignSeats(playerOrderIDs: string[], selfID: string | undefined): SeatAssignment[] {
  if (playerOrderIDs.length === 0) return [];

  const selfIdx = selfID ? playerOrderIDs.indexOf(selfID) : -1;
  const rotated =
    selfIdx === -1
      ? playerOrderIDs
      : [...playerOrderIDs.slice(selfIdx + 1), ...playerOrderIDs.slice(0, selfIdx + 1)];

  const bottom = rotated[rotated.length - 1];
  const remaining = rotated.slice(0, -1);

  const { left, top, right } = distributeAroundTable(remaining);

  return [
    { id: bottom, edge: "bottom" as const },
    ...left.map((id) => ({ id, edge: "left" as const })),
    ...top.map((id) => ({ id, edge: "top" as const })),
    ...right.map((id) => ({ id, edge: "right" as const })),
  ];
}

function distributeAroundTable(remaining: string[]): { left: string[]; top: string[]; right: string[] } {
  const n = remaining.length;
  if (n === 0) return { left: [], top: [], right: [] };
  if (n <= 2) return { left: [], top: remaining, right: [] };
  if (n === 3) return { left: [remaining[0]], top: [remaining[1]], right: [remaining[2]] };

  // n > 3: split as evenly as possible across left/top/right, giving any
  // extra seat(s) to top first (the edge opposite self) then left.
  const base = Math.floor(n / 3);
  const extra = n % 3;
  const leftCount = base + (extra === 2 ? 1 : 0);
  const topCount = base + (extra >= 1 ? 1 : 0);

  return {
    left: remaining.slice(0, leftCount),
    top: remaining.slice(leftCount, leftCount + topCount),
    right: remaining.slice(leftCount + topCount),
  };
}
