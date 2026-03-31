#!/usr/bin/env bash
set -euo pipefail

# Full-match acceptance runner for the existing single-step agent API.
# It does not require new backend business endpoints: it repeatedly calls
# /api/game/agent/run until the backend reports a winner.
#
# Usage:
#   backend/test/backend/run_agent_full_match_acceptance.sh [mock|openrouter] [model]
#
# Examples:
#   backend/test/backend/run_agent_full_match_acceptance.sh
#   backend/test/backend/run_agent_full_match_acceptance.sh openrouter qwen/qwen3.6-plus-preview:free
#
# Environment:
#   BASE_URL   default http://127.0.0.1:8080
#   MAX_STEPS  default 400

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
MODE="${1:-mock}"
MODEL="${2:-qwen/qwen3.6-plus-preview:free}"
MAX_STEPS="${MAX_STEPS:-400}"

if [[ -f "${REPO_ROOT}/.env" && -z "${OPENROUTER_API_KEY:-}" ]]; then
  # Best-effort load of local .env for manual acceptance runs.
  set +u
  set -a
  # shellcheck disable=SC1091
  source "${REPO_ROOT}/.env" >/dev/null 2>&1 || true
  set +a
  set -u
fi

if [[ "${MODE}" == "openrouter" && -z "${OPENROUTER_API_KEY:-}" ]]; then
  echo "[blocked] OPENROUTER_API_KEY is not set; openrouter acceptance cannot start." >&2
  exit 2
fi

json_get() {
  local expression="$1"
  python3 -c 'import json, sys
payload = json.load(sys.stdin)
value = payload
for part in sys.argv[1].split("."):
    if part == "":
        continue
    if isinstance(value, list):
        value = value[int(part)]
    else:
        value = value.get(part)
if isinstance(value, bool):
    print("true" if value else "false")
elif value is None:
    print("")
elif isinstance(value, (dict, list)):
    print(json.dumps(value, ensure_ascii=False))
else:
    print(value)' "${expression}"
}

get_json() {
  local path="$1"
  curl -fsS "${BASE_URL}${path}"
}

post_json() {
  local path="$1"
  local body="$2"
  curl -fsS -X POST \
    -H "Content-Type: application/json" \
    -d "${body}" \
    "${BASE_URL}${path}"
}

echo "[info] base_url=${BASE_URL}"
echo "[info] mode=${MODE}"
echo "[info] model=${MODEL}"
echo "[info] max_steps=${MAX_STEPS}"

post_json "/api/game/reset" "{}" >/dev/null

step=0
while (( step < MAX_STEPS )); do
  state_json="$(get_json "/api/game/state")"
  winner="$(printf '%s' "${state_json}" | json_get "winner")"
  if [[ -n "${winner}" ]]; then
    message="$(printf '%s' "${state_json}" | json_get "message")"
    echo "[done] winner=${winner} step=${step} message=${message}"
    exit 0
  fi

  prompt_json="$(get_json "/api/game/agent/prompt")"
  player_seat="$(printf '%s' "${prompt_json}" | json_get "playerSeat")"
  available_actions="$(printf '%s' "${prompt_json}" | json_get "availableActions")"
  counter_total="$(printf '%s' "${prompt_json}" | json_get "cardCounter.totalPlayedCardCount")"
  round_index="$(printf '%s' "${prompt_json}" | json_get "roundMemory.roundIndex")"
  trick_index="$(printf '%s' "${prompt_json}" | json_get "roundMemory.trickIndex")"

  echo "[step ${step}] seat=${player_seat} actions=${available_actions} counter.total=${counter_total} memory=${round_index}/${trick_index}"

  if [[ "${MODE}" == "openrouter" ]]; then
    request_body="$(printf '{"mode":"openrouter","model":"%s"}' "${MODEL}")"
  else
    request_body='{"mode":"mock"}'
  fi

  run_json="$(post_json "/api/game/agent/run" "${request_body}")"
  trace_error="$(printf '%s' "${run_json}" | json_get "trace.error")"
  applied="$(printf '%s' "${run_json}" | json_get "trace.applied")"
  result_message="$(printf '%s' "${run_json}" | json_get "trace.resultMessage")"
  winner="$(printf '%s' "${run_json}" | json_get "state.winner")"

  echo "[step ${step}] applied=${applied} result=${result_message}"

  if [[ -n "${trace_error}" ]]; then
    echo "[failed] trace.error=${trace_error}" >&2
    exit 1
  fi

  if [[ -n "${winner}" ]]; then
    echo "[done] winner=${winner} step=$((step + 1))"
    exit 0
  fi

  step=$((step + 1))
done

echo "[failed] no winner after ${MAX_STEPS} steps" >&2
exit 3
