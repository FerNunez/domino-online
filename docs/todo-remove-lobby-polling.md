# TODO: remove the last of the lobby-polling fallback (join/leave)

## Status

The recurring 2s poll is gone. `web/src/hooks/useLobbySnapshot.ts` (formerly
`useLobbyPolling.ts`) now does a single `getLobby` fetch on mount, and
connect/disconnect updates for already-known players arrive live over the
websocket: the gateway already relays `lobby.player_connected` /
`lobby.player_disconnected` to lobby members (its AMQP queue is bound to
`lobby.*` on the same topic exchange used for `game.*` events —
`services/api-gateway/main.go` / `shared/messaging/queue_consumer.go`), and
`web/src/hooks/useGameConnection.ts` now handles those two event types
(`LobbyEvents` in `web/src/lib/contracts.ts`), exposing a
`playerConnectivity: Record<string, boolean>` map that
`web/src/components/LobbyRoster.tsx` uses to render each player's status dot.

## What's still missing: join/leave

`lobby.player_joined` / `lobby.player_left` are declared in
`shared/contracts/amqp.go` but still never published anywhere —
`JoinLobby` (`services/lobby-service/internal/service/lobby.go`) only does
`repo.AddPlayer` and returns, with no AMQP publisher wired into that
`service` struct at all. So the one remaining `getLobby` fetch in
`useLobbySnapshot` is still load-bearing: it's the only way a client learns
about a player who joined after the page first loaded (or one who left).
Until a real join/leave event exists, that gap requires a manual
refresh/rejoin to see new players.

## When you implement join/leave events

1. Add a RabbitMQ publisher to `lobby-service`'s `service` struct, and have
   `JoinLobby`/`LeaveLobby` publish `lobby.player_joined` /
   `lobby.player_left` with enough data to render a new roster row (at
   least `userID`, `name`, `slot`; see `PlayerJoinedData`/`PlayerLeftData`
   in `shared/messaging/events.go`, already declared but unused).
2. The gateway needs no relay changes — the same `lobby.*` binding already
   used for connect/disconnect covers these too.
3. Frontend: add the two event types to `LobbyEvents`
   (`web/src/lib/contracts.ts`), handle them in `useGameConnection.ts` by
   patching a roster list held in state, and drop `useLobbySnapshot`'s
   fetch (or keep it purely as a first-paint/reconnect fallback, which is
   cheap insurance against missed events).
