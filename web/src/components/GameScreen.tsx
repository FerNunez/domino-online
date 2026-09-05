"use client";

import { useState } from "react";
import { Volume2, VolumeX } from "lucide-react";
import { Board } from "./Board";
import { Button } from "@/components/ui/button";
import { Hand } from "./Hand";
import { PlayersPanel } from "./PlayersPanel";
import { RoundHistoryModal } from "./RoundHistoryModal";
import { RoundSummary } from "./RoundSummary";
import { RoundBadge, Scoreboard } from "./Scoreboard";
import { TurnStatus } from "./TurnIndicator";
import { useSound } from "@/hooks/useSound";
import { BoardState, legalSides } from "@/lib/board";
import { LobbyModel, RoundHistoryEntry, RoundOverData, Side, Tile, slotToTeamID } from "@/lib/types";

interface GameScreenProps {
  userID: string | undefined;
  lobby: LobbyModel;
  playerOrder: string[];
  // Live connect/disconnect updates for the lifetime of this connection —
  // see useGameConnection's playerConnectivity. A player not yet present
  // here (no live event seen) falls back to the lobby snapshot's isConnected.
  playerConnectivity: Record<string, boolean>;
  handCounts: Record<string, number>;
  board: BoardState;
  hand: Tile[];
  currentTurn: string | null;
  roundOver: RoundOverData | null;
  roundHistory: RoundHistoryEntry[];
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
  playerConnectivity,
  handCounts,
  board,
  hand,
  currentTurn,
  roundOver,
  roundHistory,
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
  const { muted, toggleMute } = useSound();
  const isYourTurn = !!currentTurn && currentTurn === userID;
  const yourSlot = lobby.players.find((p) => p.id === userID)?.slot;
  const yourTeamID = yourSlot !== undefined ? slotToTeamID(yourSlot) : undefined;

  const currentTurnPlayer = currentTurn ? lobby.players.find((p) => p.id === currentTurn) : undefined;
  // Live event wins when present; otherwise fall back to the one-shot lobby
  // snapshot so a player who's been connected the whole session (no live
  // connect/disconnect event ever fired for them) doesn't show as offline.
  const currentTurnConnected = currentTurn
    ? (playerConnectivity[currentTurn] ?? currentTurnPlayer?.isConnected ?? true)
    : true;

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
    <div className="relative flex w-full max-w-5xl flex-col gap-4 p-4">
      <Button
        variant="ghost"
        size="icon"
        className="absolute right-4 top-4 z-10"
        onClick={toggleMute}
        aria-label={muted ? "Unmute sound effects" : "Mute sound effects"}
        title={muted ? "Unmute sound effects" : "Mute sound effects"}
      >
        {muted ? <VolumeX /> : <Volume2 />}
      </Button>
      {roundOver && (
        <RoundSummary
          roundOver={roundOver}
          yourTeamID={yourTeamID}
          canStartNextRound={isHost}
          nextRoundRequested={nextRoundRequested}
          pointsThisRound={latestRoundOverPoints}
          onNextRound={nextRound}
        />
      )}
      <PlayersPanel players={lobby.players} playerOrder={playerOrder} handCounts={handCounts} currentTurn={currentTurn} userID={userID} yourTeamID={yourTeamID}>
        <div className="flex flex-col items-center gap-2">
          <Board
            board={board}
            canPlayLeft={dropSides.includes("left")}
            canPlayRight={dropSides.includes("right")}
            onDropEnd={(side) => {
              if (!selectedTile) return;
              playTile(selectedTile, side);
              setSelectedTile(null);
            }}
            scoreboard={<Scoreboard gameScores={gameScores} goalScore={lobby.settings.maxScore} yourTeamID={yourTeamID} />}
            roundBadge={<RoundBadge roundNumber={roundNumber} />}
            turnStatus={
              !roundOver && (
                <TurnStatus
                  currentTurn={currentTurn}
                  currentTurnName={currentTurnPlayer?.name}
                  currentTurnConnected={currentTurnConnected}
                  userID={userID}
                />
              )
            }
            passButton={
              !roundOver && (
                <Button size="sm" variant="secondary" className="shadow-sm" disabled={!isYourTurn || hasLegalMove} onClick={pass}>
                  Pass
                </Button>
              )
            }
          />
          <RoundHistoryModal roundHistory={roundHistory} players={lobby.players} yourTeamID={yourTeamID} />
        </div>
      </PlayersPanel>
      <Hand tiles={hand} selectedTile={selectedTile} isYourTurn={isYourTurn} onSelect={handleSelect} />
    </div>
  );
}
