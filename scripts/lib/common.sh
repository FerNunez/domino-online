#!/usr/bin/env bash
# Shared config and helpers sourced by the other scripts/ files.
# Not meant to be run directly.

BASE_URL="${BASE_URL:-localhost:8081}"

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

# Extracts the userID claim straight out of a JWT.
jwt_user_id() {
    local token="$1"
    decode_jwt_payload "$(echo "$token" | cut -d. -f2)" | jq -r '.userID // empty'
}

# Extracts the lobbyID claim out of a lobby wsToken.
jwt_lobby_id() {
    local token="$1"
    decode_jwt_payload "$(echo "$token" | cut -d. -f2)" | jq -r '.lobbyID // empty'
}
