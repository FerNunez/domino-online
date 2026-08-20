import { cn } from "@/lib/utils";

interface ScoreboardProps {
  gameScores: Record<string, number>;
  goalScore: number;
  roundNumber: number;
  yourTeamID?: string;
}

// Persistent cumulative score display, mounted for the whole game (not just
// at round boundaries) — the only place a team's running total is visible.
export function Scoreboard({ gameScores, goalScore, roundNumber, yourTeamID }: ScoreboardProps) {
  const teams = Object.keys(gameScores).sort();

  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border bg-card p-3">
      <div className="flex gap-4">
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
      <p className="text-xs font-medium text-muted-foreground">Round {roundNumber}</p>
    </div>
  );
}
