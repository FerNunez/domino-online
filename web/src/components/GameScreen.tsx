"use client";

import { useState } from "react";
import { Board } from "./Board";
import { Hand } from "./Hand";
import { TurnIndicator } from "./TurnIndicator";
import { BoardState, legalSides } from "@/lib/board";
import { RoundResult, Side, Tile } from "@/lib/types";

interface GameScreenProps {
  userID: string | undefined;
  board: BoardState;
  hand: Tile[];
  currentTurn: string | null;
  roundResult: RoundResult | null;
  playTile: (tile: Tile, side: Side) => void;
  pass: () => void;
}

export function GameScreen({ userID, board, hand, currentTurn, roundResult, playTile, pass }: GameScreenProps) {
  const [selectedTile, setSelectedTile] = useState<Tile | null>(null);
  const isYourTurn = !!currentTurn && currentTurn === userID;

  const handSideOptions = new Map(hand.map((t) => [t, legalSides(board, t)]));
  const hasLegalMove = isYourTurn && Array.from(handSideOptions.values()).some((sides) => sides.length > 0);

  const handleSelect = (tile: Tile) => {
    if (!isYourTurn) return;
    const sides = legalSides(board, tile);
    if (sides.length === 0) return;
    if (sides.length === 1) {
      playTile(tile, sides[0]);
      setSelectedTile(null);
      return;
    }
    setSelectedTile(tile);
  };

  const dropSides = selectedTile ? legalSides(board, selectedTile) : [];

  return (
    <div className="flex w-full max-w-3xl flex-col gap-4 p-4">
      <TurnIndicator
        currentTurn={currentTurn}
        userID={userID}
        hasLegalMove={hasLegalMove}
        roundResult={roundResult}
        onPass={pass}
      />
      <Board
        board={board}
        canPlayLeft={dropSides.includes("left")}
        canPlayRight={dropSides.includes("right")}
        onDropEnd={(side) => {
          if (!selectedTile) return;
          playTile(selectedTile, side);
          setSelectedTile(null);
        }}
      />
      <Hand tiles={hand} selectedTile={selectedTile} isYourTurn={isYourTurn} onSelect={handleSelect} />
    </div>
  );
}
