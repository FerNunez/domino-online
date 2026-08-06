#!/usr/bin/env bash
# Opens an interactive websocat session for a guest that already joined a
# lobby (e.g. one of the guests printed by guests-join-lobby.sh), reusing
# their existing wsToken instead of joining again. Companion to
# guest-session.sh, which creates+joins+connects a brand new guest in one
# step — use this one to drive additional already-joined guests from
# separate terminals/tmux panes.
#
# Usage: ./scripts/guest-connect.sh <lobbyID> <wsToken> [baseURL]
#
# Once connected, type raw JSON client messages, e.g.:
#   {"type":"game.cmd.play_tile","data":{"tile":{"left":3,"right":5},"side":"left"}}
#   {"type":"game.cmd.pass","data":{}}
# See shared/contracts/amqp.go for the full event/command type list.
set -uo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source lib/common.sh

LOBBY_ID="${1:?usage: $0 <lobbyID> <wsToken> [baseURL]}"
WS_TOKEN="${2:?usage: $0 <lobbyID> <wsToken> [baseURL]}"
BASE_URL="${3:-$BASE_URL}"

if ! command -v websocat >/dev/null 2>&1; then
    echo "guest-connect: websocat is required but not found in PATH" >&2
    exit 1
fi

# Fail fast with a clear message instead of letting the server reject the
# upgrade with an opaque 400 (see ws.go: claims.LobbyID != lobbyID).
TOKEN_LOBBY_ID=$(jwt_lobby_id "$WS_TOKEN")
if [[ -z "$TOKEN_LOBBY_ID" ]]; then
    echo "guest-connect: couldn't decode a lobbyID claim out of that wsToken" >&2
    exit 1
fi
if [[ "$TOKEN_LOBBY_ID" != "$LOBBY_ID" ]]; then
    echo "guest-connect: this wsToken is for lobby $TOKEN_LOBBY_ID, not $LOBBY_ID" >&2
    exit 1
fi

WS_URL="ws://$BASE_URL/lobbies/$LOBBY_ID/ws?wsToken=$WS_TOKEN"

echo "lobbyID: $LOBBY_ID"
echo "websocat \"$WS_URL\""
echo "--- connected, type raw JSON messages (Ctrl-D or Ctrl-C to leave) ---"

exec websocat "$WS_URL"
