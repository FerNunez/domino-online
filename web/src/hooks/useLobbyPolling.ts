import { useEffect, useRef, useState } from "react";
import { getLobby } from "@/lib/api";
import { LobbyModel, LobbyStatus } from "@/lib/types";

const POLL_INTERVAL_MS = 2000;

// Polls GET /lobbies/{id} while the lobby is waiting for players, since the
// backend has no PlayerJoinedLobby/PlayerLeftLobby WS events (declared in
// shared/contracts/amqp.go but never published).
export function useLobbyPolling(lobbyID: string | undefined) {
  const [lobby, setLobby] = useState<LobbyModel | null>(null);
  const [error, setError] = useState<string | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (!lobbyID) return;

    let cancelled = false;

    const fetchOnce = async () => {
      try {
        const data = await getLobby(lobbyID);
        if (cancelled) return;
        setLobby(data);
        setError(null);
        if (data.status !== LobbyStatus.WAITING && intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "failed to fetch lobby");
      }
    };

    fetchOnce();
    intervalRef.current = setInterval(fetchOnce, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [lobbyID]);

  return { lobby, error };
}
