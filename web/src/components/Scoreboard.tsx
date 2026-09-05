import { cn } from "@/lib/utils";

interface ScoreboardProps {
  gameScores: Record<string, number>;
  goalScore: number;
  yourTeamID?: string;
}

// Persistent cumulative score display, mounted for the whole game (not just
// at round boundaries) — the only place a team's running total is visible.
// Plain text directly on the felt (see Board's hud slots) rather than a
// card/pill — a filled background here reads as a clickable chip, and this
// is pure readout, never interactive. The drop-shadow keeps it legible where
// tiles or the board's own tint sit behind it.
export function Scoreboard({ gameScores, goalScore, yourTeamID }: ScoreboardProps) {
  const teams = Object.keys(gameScores).sort();

  return (
    <div className="flex gap-4 [text-shadow:0_1px_2px_rgb(0_0_0/0.25)]">
      {teams.map((teamID) => {
        const isYourTeam = !!yourTeamID && teamID === yourTeamID;
        const score = gameScores[teamID] ?? 0;
        return (
          <div key={teamID} className="flex flex-col items-center">
            <p className="text-xs font-medium text-muted-foreground">
              {isYourTeam ? "Your Team" : "Opponents"}
            </p>
            <p className={cn("text-lg font-bold", isYourTeam && "text-primary")}>
              {score}
              <span className="text-xs font-normal text-muted-foreground"> / {goalScore}</span>
            </p>
          </div>
        );
      })}
    </div>
  );
}

export function RoundBadge({ roundNumber }: { roundNumber: number }) {
  return (
    <p className="text-xs font-medium text-muted-foreground [text-shadow:0_1px_2px_rgb(0_0_0/0.25)]">Round {roundNumber}</p>
  );
}
