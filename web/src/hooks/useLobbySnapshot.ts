import { useEffect, useState } from "react";
import { getLobby } from "@/lib/api";
import { LobbyModel } from "@/lib/types";

// One-shot fetch of the initial roster. Connect/disconnect updates for
// already-known players arrive live via useGameConnection's
// playerConnectivity (lobby.player_connected/disconnected relayed over the
// websocket). This does NOT catch new players joining or leaving —
// lobby-service's JoinLobby still publishes nothing (see
// docs/todo-remove-lobby-polling.md), so the roster only grows on the next
// manual page load until that's implemented.
export function useLobbySnapshot(lobbyID: string | undefined) {
  const [lobby, setLobby] = useState<LobbyModel | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!lobbyID) return;

    let cancelled = false;

    getLobby(lobbyID)
      .then((data) => {
        if (!cancelled) setLobby(data);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "failed to fetch lobby");
      });

    return () => {
      cancelled = true;
    };
  }, [lobbyID]);

  return { lobby, error };
}
