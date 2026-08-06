# Dev scripts

`lib/` holds the reusable building blocks; the top-level scripts are thin
wrappers over them. See the repo root `README.md` for the one-line summary
of each — this doc is a worked example plus gotchas.

## Tutorial: get all 4 players connected and playing

This walks through the full path from zero to 4 live WebSocket connections
(1 host + 3 guests) all in the same lobby, sending and receiving real game
events. Every step below has been run against a live stack while writing
this doc — see the root `README.md` for how to start that stack
(Postgres/Redis/RabbitMQ + the 4 Go services, or `tilt up`).

You'll use **5 terminals**: one "control" terminal to drive the HTTP calls,
and one each for the host + 3 guests' live WebSocket sessions.

### Step 1 — load the helper functions (control terminal)

```bash
source scripts/load.sh
```

This sources `lib/common.sh`, `lib/redis.sh`, and `lib/api.sh` in one call
(see the [zsh multi-file source gotcha](#gotcha-sourcing-multiple-files-in-zsh)
below for why you can't just do `source lib/common.sh lib/api.sh`). It
prints the list of functions it loaded, including `create_guest`,
`create_lobby`, `join_lobby`, and `start_lobby`.

### Step 2 — create a lobby as the host

```bash
host_token=$(create_guest | jq -r '.data.token')
lobby_resp=$(create_lobby "$host_token" 4)
lobby_id=$(echo "$lobby_resp" | sed '$d' | jq -r '.data.lobbyID')
host_ws_token=$(echo "$lobby_resp" | sed '$d' | jq -r '.data.wsToken')
echo "lobbyID: $lobby_id"
```

`create_lobby` already returns a `wsToken` for the host (no separate join
call needed for the player who created the lobby).

### Step 3 — connect the host's WebSocket (terminal 2)

Copy the printed `lobbyID` into a **second terminal** (shell variables
don't cross terminals — see the
[tty gotcha](#gotcha-guest-sessionsh-needs-its-own-terminal) below) and
either:

```bash
# reuses the host's own identity/token instead of creating a fresh guest:
./scripts/guest-session.sh "<lobbyID>" localhost:8081 "$host_token"
```

or, if you'd rather stay in the control terminal and paste the token
directly:

```bash
./scripts/guest-connect.sh "$lobby_id" "$host_ws_token"
```

You're now connected live as the host — you'll see `lobby.player_joined`
events land as the other 3 guests join in the next step.

### Step 4 — join 3 more guests (control terminal)

```bash
./scripts/guests-join-lobby.sh "$lobby_id" localhost:8081 3
```

This creates 3 fresh guests, joins them all to the lobby over plain HTTP,
and prints each one's `userID`/`wsToken` plus a ready-to-paste `websocat`
command — but it does **not** open a WebSocket for them itself.

### Step 5 — connect each guest's WebSocket (terminals 3, 4, 5)

In three more separate terminals/panes, one per guest, using the
`wsToken` `guests-join-lobby.sh` printed for that guest:

```bash
./scripts/guest-connect.sh "$lobby_id" "<guest's wsToken>"
```

All 4 terminals are now independent, live send/receive JSON sessions
against the same lobby.

### Step 6 — start the game (control terminal)

```bash
start_lobby "$host_token" "$lobby_id"
```

Once all 4 are connected, this fans out `game.game_started` (with
`currentTurn`/`playerOrder`) to every socket, plus a `game.hand_dealt`
event to each player carrying *their own* 7 tiles. From here, whichever
player's turn it is can send:

```
{"type":"game.cmd.play_tile","data":{"tile":{"left":3,"right":5},"side":"left"}}
{"type":"game.cmd.pass","data":{}}
```

and every connected socket sees the resulting `game.move_made` /
`game.player_passed` / `game.turn_changed` events land in real time.

If something gets stuck mid-test, reset it and rerun without restarting
services:

```bash
./scripts/redis-cli.sh reset-lobby-status "$lobby_id"
```

### Step 7 — look up a lobby's raw state (control terminal)

To inspect a lobby's current state directly in Redis (Players/Settings/Status/etc),
by its `lobbyID`:

```bash
domino-redis-cli GET "lobby:$lobby_id" | jq .
```

`domino-redis-cli` pulls the live `REDIS_URI` from the k8s secret on every
call, so it's never stale. If you don't have the exact ID handy, scan and
pretty-print every lobby key at once instead:

```bash
domino-redis-keys-values "lobby:*"
```

## Gotcha: `guest-session.sh` needs its own terminal

`guest-session.sh` execs into `websocat`, which reads interactively from
the terminal (a tty). Backgrounding it with `&` in the same shell as
another foreground command doesn't work — the moment `websocat` tries to
read stdin it gets stopped (`suspended (tty input)`), because the
foreground command still owns the terminal. Run it in its own
terminal/pane instead of backgrounding it. If it's already stuck
suspended, `fg` brings it back so you can type into it.

Relatedly: shell variables like `$lobby_id` don't cross into a new
terminal — each terminal has its own environment. Either copy the actual
ID printed by `create_lobby`/`guest-session.sh`'s banner and paste it
literally into the other terminal, or `export lobby_id` if the second
terminal is a child shell of the first (a new tab/window usually isn't).

## Gotcha: sourcing multiple files in zsh

`source a.sh b.sh` does **not** source both files in zsh (or bash) — only
`a.sh` is sourced; `b.sh` becomes `$1`, a positional parameter passed into
`a.sh`. Source each file on its own line/command instead:

```bash
source scripts/lib/common.sh
source scripts/lib/api.sh
# or: source scripts/lib/common.sh; source scripts/lib/api.sh
```

The scripts themselves (`guest-session.sh`, `guests-join-lobby.sh`,
`redis-cli.sh`) aren't affected — each sources its own deps internally,
one `source` call per file.

## Other capabilities

- **`start_lobby(token, lobbyID)`** in `lib/api.sh` — kick off the game
  once the lobby's full, same pattern as `create_lobby`/`join_lobby`
  (body + `%{http_code}` on the last line).
- **Token reuse** — `guest-session.sh <lobbyID> <baseURL> <token>` accepts
  a token as a 3rd arg, so you can reconnect the *same* simulated guest
  after a dropped connection instead of getting a fresh identity.
- **`guest-connect.sh <lobbyID> <wsToken> [baseURL]`** — connects with an
  existing `wsToken` directly, no join call. Use this for guests that
  already joined (see the [tutorial](#tutorial-get-all-4-players-connected-and-playing)
  above).
- **`jwt_user_id "$token"`** — decode any token's `userID` claim on the
  fly, handy for correlating log lines to a guest you spun up.
- **`BASE_URL` env override** — every script respects `BASE_URL`, so you
  can point at a Tilt/k8s-hosted gateway instead of `localhost:8081`
  without editing anything:
  `BASE_URL=my-cluster:8081 ./scripts/guest-session.sh <lobbyID>`.
- **Multiple watchers on one lobby** — nothing stops you running
  `guest-session.sh`/`guest-connect.sh` in several terminals against the
  same `lobbyID` to eyeball fan-out (e.g. confirming all players get
  `game.turn_changed`).

## Known gap

There's no scripted "play a full game end-to-end" driver yet — each
`game.cmd.*` message is still hand-typed. A `play-random-game.sh` that
loops guest turns automatically would be a natural next addition.
