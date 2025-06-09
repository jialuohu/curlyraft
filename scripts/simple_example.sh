#!/usr/bin/env bash
set -euo pipefail

STORAGE_DIR="${1:-storage}"
CMD_DIR="${2}"
mkdir -p "$STORAGE_DIR"

# list of child PIDs
pids=()

# on exit (Ctrl+C or kill), shut down all children
cleanup() {
  echo "Shutting down servers…"
  kill "${pids[@]}" 2>/dev/null
  wait "${pids[@]}" 2>/dev/null
  echo "All servers stopped."
  exit
}
trap cleanup SIGINT SIGTERM

# start nodeA
go run "$CMD_DIR/main.go" \
  --id    nodeA \
  --addr  localhost:21001 \
  --peers nodeB/localhost:21002,nodeC/localhost:21003 \
  --data  "$STORAGE_DIR/21001" &
pids+=($!)

# start nodeB
go run "$CMD_DIR/main.go" \
  --id    nodeB \
  --addr  localhost:21002 \
  --peers nodeA/localhost:21001,nodeC/localhost:21003 \
  --data  "$STORAGE_DIR/21002" &
pids+=($!)

# start nodeC
go run "$CMD_DIR/main.go" \
  --id    nodeC \
  --addr  localhost:21003 \
  --peers nodeA/localhost:21001,nodeB/localhost:21002 \
  --data  "$STORAGE_DIR/21003" &
pids+=($!)

echo "Started servers with PIDs: ${pids[*]}"
echo "Press Ctrl+C to stop them."

# wait for all of them (or for the trap)
wait
