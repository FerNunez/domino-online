"use client";

import { useEffect, useMemo, useState } from "react";
import { RoundActionsList } from "./RoundActionsList";
import { ApiError, getRoundActions } from "@/lib/api";
import { HistoryAction, HistoryHand, PlayerModel } from "@/lib/types";

interface RoundActionsPanelProps {
  roundID: string;
  players: PlayerModel[];
}

// history-service persists a round asynchronously (over RabbitMQ) after
// RoundOver reaches this client over WS, so the round's row — and therefore
// its actions/hands — may not exist in Postgres yet the instant this panel
// mounts. GetRoundActions returns 404 (ApiError.status) until it does; these
// are the retry delays between attempts before giving up. See grpc_handler.go's
// GetRoundActions: a round's actions/hands are always fully persisted before
// its `rounds` row is written, so once a fetch succeeds it's guaranteed
// complete — never a partial list caught mid-persist.
const RETRY_DELAYS_MS = [0, 400, 800, 1500, 3000];

type PanelState =
  | { status: "loading" }
  | { status: "ready"; actions: HistoryAction[]; hands: HistoryHand[] }
  | { status: "unavailable" };

// Auto-fetches and displays this round's move-by-move actions + starting
// hands right after it ends, so the live game screen shows the same detail
// GameRoundRow shows for past games — without the player navigating away.
export function RoundActionsPanel({ roundID, players }: RoundActionsPanelProps) {
  const [state, setState] = useState<PanelState>({ status: "loading" });

  const playerName = useMemo(() => {
    const byID = new Map(players.map((p) => [p.id, p.name]));
    return (id: string) => byID.get(id) ?? `${id.slice(0, 8)}…`;
  }, [players]);

  useEffect(() => {
    let cancelled = false;
    setState({ status: "loading" });

    (async () => {
      for (let attempt = 0; attempt < RETRY_DELAYS_MS.length; attempt++) {
        if (attempt > 0) {
          await new Promise((resolve) => setTimeout(resolve, RETRY_DELAYS_MS[attempt]));
        }
        if (cancelled) return;
        try {
          const data = await getRoundActions(roundID);
          if (!cancelled) setState({ status: "ready", actions: data.actions, hands: data.hands });
          return;
        } catch (err) {
          const notReadyYet = err instanceof ApiError && err.status === 404;
          if (!notReadyYet) {
            if (!cancelled) setState({ status: "unavailable" });
            return;
          }
          // 404 — round not persisted yet, fall through and retry.
        }
      }
      if (!cancelled) setState({ status: "unavailable" });
    })();

    return () => {
      cancelled = true;
    };
  }, [roundID]);

  if (state.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading this round&apos;s moves…</p>;
  }

  if (state.status === "unavailable") {
    return (
      <p className="text-sm text-muted-foreground">
        Still saving this round&apos;s history — you&apos;ll find it under My Games shortly.
      </p>
    );
  }

  return <RoundActionsList actions={state.actions} hands={state.hands} playerName={playerName} />;
}
