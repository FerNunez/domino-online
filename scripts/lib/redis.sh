#!/usr/bin/env bash
# Redis helpers for the domino dev cluster.
# Not meant to be run directly — see scripts/redis-cli.sh for the CLI entrypoint.
# NOTE: no top-level `set -u`/`set -e` here — this file is meant to be
# sourced into an interactive shell, and shell options set outside a
# function leak into the caller's shell (breaking prompts that reference
# unset variables).

# Pulls the current REDIS_URI from the k8s secret so it never goes stale/hardcoded.
domino-redis-uri() {
  kubectl get secret redis -o jsonpath='{.data.uri}' | base64 -d
}

domino-redis-cli() {
  redis-cli -u "$(domino-redis-uri)" "$@"
}

# Resets a lobby's Status field in place, preserving Players/Settings/etc.
# Usage: reset-lobby-status <lobbyID> [status]
#   status defaults to LOBBY_STATUS_WAITING
reset-lobby-status() {
  local lobby_id="${1:?usage: reset-lobby-status <lobbyID> [status]}"
  local new_status="${2:-LOBBY_STATUS_WAITING}"
  local key="lobby:$lobby_id"

  local current
  current=$(domino-redis-cli GET "$key")
  if [[ -z "$current" || "$current" == "(nil)" ]]; then
    echo "reset-lobby-status: no lobby found for key $key" >&2
    return 1
  fi

  local updated
  updated=$(echo "$current" | jq -c --arg status "$new_status" '.Status = $status')

  domino-redis-cli SET "$key" "$updated" >/dev/null
  echo "lobby $lobby_id -> $new_status"
}

# Scans for keys matching a pattern and GETs each one, pretty-printing
# JSON values via jq (falling back to raw output for non-JSON values).
# A `while read` loop, not `xargs` — xargs execs its target as a program
# via PATH and can't see shell functions like domino-redis-cli.
# Piping the whole thing through `jq` yourself won't work: the "== key =="
# headers between values aren't valid JSON, so this formats each value
# individually instead.
# Usage: domino-redis-keys-values <pattern>
#   e.g. domino-redis-keys-values "lobby:*"
domino-redis-keys-values() {
  local pattern="${1:?usage: domino-redis-keys-values <pattern>}"
  local found=0 key value

  while read -r key; do
    [[ -z "$key" ]] && continue
    found=1
    echo "== $key =="
    value=$(domino-redis-cli GET "$key")
    if echo "$value" | jq -e . >/dev/null 2>&1; then
      echo "$value" | jq .
    else
      echo "$value"
    fi
  done < <(domino-redis-cli --scan --pattern "$pattern")

  if [[ "$found" -eq 0 ]]; then
    echo "domino-redis-keys-values: no keys matched $pattern" >&2
    return 1
  fi
}
