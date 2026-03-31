#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="$ROOT_DIR/.run"

BACKEND_PORT=8080
FRONTEND_PORT=5173

BACKEND_PID_FILE="$RUN_DIR/backend.pid"
FRONTEND_PID_FILE="$RUN_DIR/frontend.pid"
BACKEND_LOG="$RUN_DIR/backend.log"
FRONTEND_LOG="$RUN_DIR/frontend.log"

mkdir -p "$RUN_DIR"

kill_pid_file() {
  local pid_file="$1"
  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(cat "$pid_file" 2>/dev/null || true)"
    if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      sleep 1
      kill -9 "$pid" 2>/dev/null || true
    fi
    rm -f "$pid_file"
  fi
}

kill_port() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids="$(lsof -ti tcp:"$port" || true)"
    if [[ -n "${pids:-}" ]]; then
      kill $pids 2>/dev/null || true
      sleep 1
      kill -9 $pids 2>/dev/null || true
    fi
    return
  fi

  if command -v fuser >/dev/null 2>&1; then
    fuser -k "${port}/tcp" 2>/dev/null || true
  fi
}

wait_for_http() {
  local url="$1"
  local name="$2"

  for _ in {1..60}; do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done

  echo "[restart] ${name} did not become ready: ${url}" >&2
  return 1
}

start_process() {
  local pid_file="$1"
  local log_file="$2"
  shift 2

  nohup setsid "$@" </dev/null >"$log_file" 2>&1 &
  echo "$!" >"$pid_file"
}

kill_pid_file "$BACKEND_PID_FILE"
kill_pid_file "$FRONTEND_PID_FILE"
kill_port "$BACKEND_PORT"
kill_port "$FRONTEND_PORT"

cd "$ROOT_DIR"

start_process "$BACKEND_PID_FILE" "$BACKEND_LOG" npm run backend:dev
start_process "$FRONTEND_PID_FILE" "$FRONTEND_LOG" npm run dev -- --host 127.0.0.1

wait_for_http "http://127.0.0.1:${BACKEND_PORT}/api/game/state" "backend"
wait_for_http "http://127.0.0.1:${FRONTEND_PORT}" "frontend"

echo "[restart] backend  pid: $(cat "$BACKEND_PID_FILE")"
echo "[restart] frontend pid: $(cat "$FRONTEND_PID_FILE")"
echo "[restart] backend  url: http://127.0.0.1:${BACKEND_PORT}/api/game/state"
echo "[restart] frontend url: http://127.0.0.1:${FRONTEND_PORT}"
echo "[restart] backend  log: $BACKEND_LOG"
echo "[restart] frontend log: $FRONTEND_LOG"
