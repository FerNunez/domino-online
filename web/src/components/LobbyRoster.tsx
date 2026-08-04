import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { LobbyModel } from "@/lib/types";

interface LobbyRosterProps {
  lobby: LobbyModel;
}

export function LobbyRoster({ lobby }: LobbyRosterProps) {
  const slots = Array.from({ length: lobby.maxPlayers }, (_, i) => lobby.players.find((p) => p.slot === i + 1));

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
      {slots.map((player, i) => (
        <Card key={i} className={cn("border-2", player ? "border-primary/40" : "border-dashed")}>
          <CardContent className="flex flex-col items-center gap-1 p-4 text-center">
            <div
              className={cn(
                "flex h-12 w-12 items-center justify-center rounded-full text-lg font-semibold",
                player ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
              )}
            >
              {player ? player.name.slice(0, 2).toUpperCase() : "?"}
            </div>
            <p className="truncate text-sm font-medium">{player ? player.name : "Waiting…"}</p>
            {player && player.id === lobby.hostId && (
              <span className="rounded-full bg-secondary px-2 py-0.5 text-xs text-secondary-foreground">Host</span>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
