#!/bin/sh
set -u

child=""

stop() {
  if [ -n "$child" ]; then
    kill -TERM "$child" 2>/dev/null || true
    wait "$child" 2>/dev/null || true
  fi
  exit 0
}

trap stop INT TERM

while :; do
  /usr/local/bin/app "$@" &
  child=$!
  wait "$child"
  status=$?
  child=""

  if [ "$status" -ne 0 ]; then
    echo "node process exited with status $status; retrying in 2 seconds" >&2
  fi

  sleep 2 &
  child=$!
  wait "$child"
  child=""
done