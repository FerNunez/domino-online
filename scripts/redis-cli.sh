#!/usr/bin/env bash
# CLI entrypoint over scripts/lib/redis.sh.
# Usage:
#   source scripts/lib/redis.sh          # load functions into your shell
#   ./scripts/redis-cli.sh reset-lobby-status <lobbyID> [status]
set -uo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source lib/redis.sh

cmd="${1:-}"
case "$cmd" in
    reset-lobby-status)
        shift
        reset-lobby-status "$@"
        ;;
    "")
        echo "usage: $0 reset-lobby-status <lobbyID> [status]" >&2
        exit 1
        ;;
    *)
        echo "unknown command: $cmd" >&2
        exit 1
        ;;
esac
