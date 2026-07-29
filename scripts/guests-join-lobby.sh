#!/usr/bin/env bash
# Usage: ./scripts/guests-join-lobby.sh <lobbyID> [baseURL] [numGuests]
set -uo pipefail

LOBBY_ID="${1:?usage: $0 <lobbyID> [baseURL] [numGuests]}"
BASE_URL="${2:-localhost:8081}"
NUM_GUESTS="${3:-3}"

# Decodes a JWT payload segment: JWTs use base64url (RFC 4648 §5) with
# padding stripped, which plain `base64 -d` doesn't understand.
decode_jwt_payload() {
    local segment="${1//-/+}"
    segment="${segment//_//}"
    case $(( ${#segment} % 4 )) in
        2) segment+="==" ;;
        3) segment+="=" ;;
    esac
    echo "$segment" | base64 -d 2>/dev/null
}

create_guest() {
    curl -s -X POST "$BASE_URL/auth/guest"
}

join_lobby() {
    local token="$1"
    curl -s -w '\n%{http_code}' -X POST "$BASE_URL/lobbies/$LOBBY_ID/join" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json"
}

declare -a user_ids
declare -a tokens

for i in $(seq 1 "$NUM_GUESTS"); do
    guest_resp=$(create_guest)
    token=$(echo "$guest_resp" | jq -r '.data.token // empty')

    if [[ -z "$token" ]]; then
        echo "guest $i: failed to create guest -> $guest_resp"
        continue
    fi

    user_id=$(decode_jwt_payload "$(echo "$token" | cut -d. -f2)" | jq -r '.userID // empty')

    tokens+=("$token")
    user_ids+=("$user_id")
done

for idx in "${!tokens[@]}"; do
    token="${tokens[$idx]}"
    user_id="${user_ids[$idx]}"

    resp=$(join_lobby "$token")
    http_code=$(echo "$resp" | tail -n1)
    body=$(echo "$resp" | sed '$d')

    if [[ "$http_code" -ge 200 && "$http_code" -lt 300 ]]; then
        ws_token=$(echo "$body" | jq -r '.data.wsToken // empty')
        echo "userID: $user_id  wsToken: $ws_token"
        echo "websocat \"ws://$BASE_URL/lobbies/$LOBBY_ID/ws?wsToken=$ws_token\""
        echo
    else
        echo "userID: $user_id -> error (HTTP $http_code): $body"
    fi
done
