import { ReactNode } from "react";
import { PlayerSeat } from "./PlayerSeat";
import { assignSeats, SeatEdge } from "@/lib/seatLayout";
import { PlayerModel } from "@/lib/types";

interface PlayersPanelProps {
  players: PlayerModel[];
  playerOrder: string[];
  handCounts: Record<string, number>;
  currentTurn: string | null;
  userID: string | undefined;
  children: ReactNode;
}

// Renders the seat ring as a CSS grid with the board (passed as `children`)
// occupying the center cell — this component only knows about seats, never
// about Board, keeping the two fully decoupled.
export function PlayersPanel({ players, playerOrder, handCounts, currentTurn, userID, children }: PlayersPanelProps) {
  const seats = assignSeats(playerOrder, userID);
  const byEdge: Record<SeatEdge, typeof seats> = { bottom: [], left: [], top: [], right: [] };
  for (const seat of seats) byEdge[seat.edge].push(seat);

  const renderEdge = (edge: SeatEdge, direction: "row" | "col") => (
    <div className={`flex items-center justify-center gap-2 ${direction === "col" ? "flex-col" : ""}`} style={{ gridArea: edge }}>
      {byEdge[edge].map((seat) => (
        <PlayerSeat
          key={seat.id}
          player={players.find((p) => p.id === seat.id)}
          handCount={handCounts[seat.id] ?? 0}
          isCurrentTurn={seat.id === currentTurn}
          isSelf={seat.id === userID}
        />
      ))}
    </div>
  );

  return (
    <div
      className="grid w-full gap-3"
      style={{
        gridTemplateAreas: '"left top right" "left center right" "left bottom right"',
        gridTemplateColumns: "auto 1fr auto",
        gridTemplateRows: "auto 1fr auto",
      }}
    >
      {renderEdge("left", "col")}
      {renderEdge("top", "row")}
      {renderEdge("right", "col")}
      <div style={{ gridArea: "center" }}>{children}</div>
      {renderEdge("bottom", "row")}
    </div>
  );
}
