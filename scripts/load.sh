#!/usr/bin/env bash
# Sources every scripts/lib/*.sh file into your current shell in one go,
# so you don't hit the "source a.sh b.sh only sources a.sh" gotcha.
#
# Usage (from anywhere, bash or zsh):
#   source /path/to/scripts/load.sh
#
# Not meant to be executed directly — it defines functions
# (create_guest, join_lobby, domino-redis-cli, jwt_user_id, ...) that only
# make sense in your current interactive shell.

if [ -n "${ZSH_VERSION:-}" ]; then
    _load_src="${(%):-%x}"
else
    _load_src="${BASH_SOURCE[0]}"
fi
_load_dir="$(cd "$(dirname "$_load_src")" && pwd)"

source "$_load_dir/lib/common.sh"
source "$_load_dir/lib/redis.sh"
source "$_load_dir/lib/api.sh"

unset _load_src _load_dir

echo "domino dev scripts loaded: create_guest, create_lobby, join_lobby, start_lobby, jwt_user_id, decode_jwt_payload, domino-redis-uri, domino-redis-cli, reset-lobby-status"
