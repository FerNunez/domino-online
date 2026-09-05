"use client";

import { useEffect, useState } from "react";
import { DominoTile } from "./DominoTile";
import { assignSeats, SeatEdge } from "@/lib/seatLayout";
import { PlayerModel, Tile, slotToTeamID, tileToString } from "@/lib/types";
import { cn } from "@/lib/utils";

interface RoundHandsRevealProps {
  // Every player's remaining tiles at round end (RoundOverData.playerHands).
  // Undefined on the rare reconnect-snapshot path — render nothing then,
  // rather than a reveal with missing hands.
  playerHands: Record<string, Tile[]> | undefined;
  players: PlayerModel[];
  playerOrder: string[];
  userID: string | undefined;
  yourTeamID?: string;
}

// Starting transform for a tile's slide-in, keyed by the player's actual
// seat edge (see seatLayout) — each hand drifts in from roughly the
// direction its owner sits at, a cheap stand-in for a real from-seat FLIP
// animation that doesn't need cross-component DOM measurement.
const EDGE_START: Record<SeatEdge, string> = {
  top: "-translate-y-4 opacity-0",
  bottom: "translate-y-4 opacity-0",
  left: "-translate-x-4 opacity-0",
  right: "translate-x-4 opacity-0",
};

// Sits directly on the felt (see Board's `reveal` slot) rather than in a
// popup — grouped by team, then by player, using the same "sm" tile size as
// the board itself so revealed hands read as part of the table, not an
// overlay on top of it.
export function RoundHandsReveal({ playerHands, players, playerOrder, userID, yourTeamID }: RoundHandsRevealProps) {
  const [entered, setEntered] = useState(false);

  useEffect(() => {
    const id = requestAnimationFrame(() => setEntered(true));
    return () => cancelAnimationFrame(id);
  }, []);

  if (!playerHands) return null;

  const seatByID = new Map(assignSeats(playerOrder, userID).map((s) => [s.id, s.edge]));
  const teamIDs = Array.from(new Set(players.map((p) => slotToTeamID(p.slot)))).sort((a, b) =>
    a === yourTeamID ? -1 : b === yourTeamID ? 1 : 0
  );

  return (
    <div className="pointer-events-none absolute inset-x-0 bottom-4 flex max-w-full flex-wrap items-end justify-center gap-6 px-4">
      {teamIDs.map((teamID) => (
        <div key={teamID} className="flex max-w-full flex-col items-center gap-1.5 rounded-lg bg-background/85 p-2 backdrop-blur-sm">
          <p className="text-[11px] font-medium text-muted-foreground">{teamID === yourTeamID ? "Your team" : "Opponents"}</p>
          <div className="flex max-w-full flex-wrap items-start justify-center gap-x-4 gap-y-2">
            {players
              .filter((p) => slotToTeamID(p.slot) === teamID)
              .map((player) => {
                const tiles = playerHands[player.id] ?? [];
                const edge = seatByID.get(player.id) ?? "top";
                return (
                  <div key={player.id} className="flex flex-col items-center gap-1">
                    <p className="text-[10px] text-muted-foreground">{player.name}</p>
                    {tiles.length === 0 ? (
                      <p className="text-[10px] font-medium">domino!</p>
                    ) : (
                      <div className="flex max-w-[220px] flex-wrap justify-center gap-1">
                        {tiles.map((tile, i) => (
                          <div
                            key={`${tileToString(tile)}-${i}`}
                            className={cn("transition-all duration-500 ease-out", entered ? "translate-x-0 translate-y-0 opacity-100" : EDGE_START[edge])}
                            style={{ transitionDelay: `${i * 60}ms` }}
                          >
                            <DominoTile tile={tile} size="sm" />
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
          </div>
        </div>
      ))}
    </div>
  );
}
