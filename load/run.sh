#!/bin/sh
set -eu

export START_RPS=1 RAMP_DURATION=2m HOLD_DURATION=1h
pandora_pid=

stop_pandora() {
    if [ -n "$pandora_pid" ]; then
        kill -TERM "$pandora_pid" 2>/dev/null || true
        wait "$pandora_pid" || true
    fi
    exit 0
}

trap stop_pandora TERM INT

while true; do
    pandora /load/load.yaml &
    pandora_pid=$!
    wait "$pandora_pid"
    pandora_pid=
    export START_RPS=20 RAMP_DURATION=1s
done
