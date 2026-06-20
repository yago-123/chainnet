#!/bin/sh
set -eu

if [ -n "${CONFIG_FILE:-}" ]; then
	exec /usr/local/bin/chainnet --config "$CONFIG_FILE" "$@"
fi

exec /usr/local/bin/chainnet "$@"
