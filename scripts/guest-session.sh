#!/usr/bin/env bash
# Simulates a guest joining a lobby and holds an interactive websocat
# session open against it, so you can watch server events and hand-type
# client commands live instead of copy-pasting a websocat invocation.
#
# Usage: ./scripts/guest-session.sh <lobbyID> [baseURL] [token]
#   lobbyID  - required
#   baseURL  - defaults to $BASE_URL / localhost:8081
#   token    - an existing guest/user JWT to reuse; if omitted a fresh
#              guest is created
#
# Once connected, type raw JSON client messages, e.g.:
#   {"type":"game.cmd.play_tile","data":{"tile":{"left":3,"right":5},"side":"left"}}
#   {"type":"game.cmd.pass","data":{}}
# See shared/contracts/amqp.go for the full event/command list.
set -uo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source lib/common.sh
source lib/api.sh

LOBBY_ID="${1:?usage: $0 <lobbyID> [baseURL] [token]}"
BASE_URL="${2:-$BASE_URL}"
TOKEN="${3:-}"

if ! command -v websocat >/dev/null 2>&1; then
    echo "guest-session: websocat is required but not found in PATH" >&2
    exit 1
fi

if [[ -z "$TOKEN" ]]; then
    guest_resp=$(create_guest)
    TOKEN=$(echo "$guest_resp" | jq -r '.data.token // empty')
    if [[ -z "$TOKEN" ]]; then
        echo "guest-session: failed to create guest -> $guest_resp" >&2
        exit 1
    fi
fi

USER_ID=$(jwt_user_id "$TOKEN")

resp=$(join_lobby "$TOKEN" "$LOBBY_ID")
http_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | sed '$d')

if [[ "$http_code" -lt 200 || "$http_code" -ge 300 ]]; then
    echo "guest-session: failed to join lobby $LOBBY_ID (HTTP $http_code): $body" >&2
    exit 1
fi

WS_TOKEN=$(echo "$body" | jq -r '.data.wsToken // empty')
if [[ -z "$WS_TOKEN" ]]; then
    echo "guest-session: joined but no wsToken in response: $body" >&2
    exit 1
fi

WS_URL="ws://$BASE_URL/lobbies/$LOBBY_ID/ws?wsToken=$WS_TOKEN"

echo "userID:  $USER_ID"
echo "lobbyID: $LOBBY_ID"
echo "websocat \"$WS_URL\""
echo "--- connected, type raw JSON messages (Ctrl-D or Ctrl-C to leave) ---"

exec websocat "$WS_URL"
