import { cn } from "@/lib/utils";

interface ScoreboardProps {
  gameScores: Record<string, number>;
  goalScore: number;
  yourTeamID?: string;
}

// Persistent cumulative score display, mounted for the whole game (not just
// at round boundaries) — the only place a team's running total is visible.
// Rendered as a compact badge overlaid on the board (see Board's hud slots),
// so it stays a tight inline cluster rather than a full-width bar.
export function Scoreboard({ gameScores, goalScore, yourTeamID }: ScoreboardProps) {
  const teams = Object.keys(gameScores).sort();

  return (
    <div className="flex gap-4 rounded-md bg-card/90 px-3 py-1.5 shadow-sm">
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
    <p className="rounded-md bg-card/90 px-3 py-1.5 text-xs font-medium text-muted-foreground shadow-sm">
      Round {roundNumber}
    </p>
  );
}
