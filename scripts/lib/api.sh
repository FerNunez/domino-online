#!/usr/bin/env bash
# api-gateway HTTP helpers (see services/api-gateway/main.go for routes).
# Not meant to be run directly. Requires scripts/lib/common.sh to be
# sourced first for $BASE_URL.
#
# join_lobby / create_lobby / start_lobby print "<body>\n<http_code>" so
# callers can split the last line off with `tail -n1` / `sed '$d'`, same
# as the pattern already used in guests-join-lobby.sh.

create_guest() {
    curl -s -X POST "$BASE_URL/auth/guest"
}

create_lobby() {
    local token="$1" max_players="$2"
    curl -s -w '\n%{http_code}' -X POST "$BASE_URL/lobbies" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{\"maxPlayers\":$max_players}"
}

join_lobby() {
    local token="$1" lobby_id="$2"
    curl -s -w '\n%{http_code}' -X POST "$BASE_URL/lobbies/$lobby_id/join" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json"
}

start_lobby() {
    local token="$1" lobby_id="$2"
    curl -s -w '\n%{http_code}' -X POST "$BASE_URL/lobbies/$lobby_id/start" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json"
}
