"use client";

import { useState } from "react";
import { Board } from "./Board";
import { Hand } from "./Hand";
import { PlayersPanel } from "./PlayersPanel";
import { RoundSummary } from "./RoundSummary";
import { Scoreboard } from "./Scoreboard";
import { TurnIndicator } from "./TurnIndicator";
import { BoardState, legalSides } from "@/lib/board";
import { LobbyModel, RoundOverData, Side, Tile, slotToTeamID } from "@/lib/types";

interface GameScreenProps {
  userID: string | undefined;
  lobby: LobbyModel;
  playerOrder: string[];
  handCounts: Record<string, number>;
  board: BoardState;
  hand: Tile[];
  currentTurn: string | null;
  roundOver: RoundOverData | null;
  gameScores: Record<string, number>;
  roundNumber: number;
  latestRoundOverPoints: Record<string, number> | null;
  nextRoundRequested: boolean;
  isHost: boolean;
  playTile: (tile: Tile, side: Side) => void;
  pass: () => void;
  nextRound: () => void;
}

export function GameScreen({
  userID,
  lobby,
  playerOrder,
  handCounts,
  board,
  hand,
  currentTurn,
  roundOver,
  gameScores,
  roundNumber,
  latestRoundOverPoints,
  nextRoundRequested,
  isHost,
  playTile,
  pass,
  nextRound,
}: GameScreenProps) {
  const [selectedTile, setSelectedTile] = useState<Tile | null>(null);
  const isYourTurn = !!currentTurn && currentTurn === userID;
  const yourSlot = lobby.players.find((p) => p.id === userID)?.slot;
  const yourTeamID = yourSlot !== undefined ? slotToTeamID(yourSlot) : undefined;

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
    <div className="flex w-full max-w-5xl flex-col gap-4 p-4">
      <Scoreboard gameScores={gameScores} goalScore={lobby.settings.maxScore} roundNumber={roundNumber} yourTeamID={yourTeamID} />
      {roundOver ? (
        <RoundSummary
          roundOver={roundOver}
          yourTeamID={yourTeamID}
          canStartNextRound={isHost}
          nextRoundRequested={nextRoundRequested}
          pointsThisRound={latestRoundOverPoints}
          onNextRound={nextRound}
        />
      ) : (
        <TurnIndicator currentTurn={currentTurn} userID={userID} hasLegalMove={hasLegalMove} onPass={pass} />
      )}
      <PlayersPanel players={lobby.players} playerOrder={playerOrder} handCounts={handCounts} currentTurn={currentTurn} userID={userID} yourTeamID={yourTeamID}>
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
      </PlayersPanel>
      <Hand tiles={hand} selectedTile={selectedTile} isYourTurn={isYourTurn} onSelect={handleSelect} />
    </div>
  );
}
