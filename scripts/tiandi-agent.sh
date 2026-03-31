#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${TIANDI_BASE_URL:-http://127.0.0.1:8080}"

usage() {
  cat <<'EOF'
Usage:
  scripts/tiandi-agent.sh [--base-url URL] state
  scripts/tiandi-agent.sh [--base-url URL] reset
  scripts/tiandi-agent.sh [--base-url URL] prompt
  scripts/tiandi-agent.sh [--base-url URL] trace
  scripts/tiandi-agent.sh [--base-url URL] match
  scripts/tiandi-agent.sh [--base-url URL] match-state
  scripts/tiandi-agent.sh [--base-url URL] match-trace
  scripts/tiandi-agent.sh [--base-url URL] match-reset
  scripts/tiandi-agent.sh [--base-url URL] action --seat P0 --kind play [--cards D3,S3] [--resolution-id candidate-1]
  scripts/tiandi-agent.sh [--base-url URL] action '{"seat":"P0","kind":"play","cards":["D3"]}'
  scripts/tiandi-agent.sh [--base-url URL] run-mock [--model MODEL]
  scripts/tiandi-agent.sh [--base-url URL] run-openrouter [--model MODEL]
  scripts/tiandi-agent.sh [--base-url URL] match-run [--model MODEL] [--max-steps N] [--fallback-mode mock]
  scripts/tiandi-agent.sh [--base-url URL] match-run-mock [--max-steps N]

Environment:
  TIANDI_BASE_URL=http://127.0.0.1:8080
EOF
}

json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\n'/\\n}"
  printf '%s' "$value"
}

json_array_from_csv() {
  local csv="$1"
  local items=()
  local item
  local i

  if [[ -z "$csv" ]]; then
    printf '[]'
    return
  fi

  IFS=',' read -r -a items <<<"$csv"
  printf '['
  for i in "${!items[@]}"; do
    item="${items[$i]}"
    item="${item#"${item%%[![:space:]]*}"}"
    item="${item%"${item##*[![:space:]]}"}"
    printf '"%s"' "$(json_escape "$item")"
    if (( i + 1 < ${#items[@]} )); then
      printf ','
    fi
  done
  printf ']'
}

print_response() {
  if command -v jq >/dev/null 2>&1; then
    jq .
  else
    cat
  fi
}

curl_json() {
  local method="$1"
  local path="$2"
  local body="${3-}"

  if [[ -n "$body" ]]; then
    curl --silent --show-error \
      -X "$method" \
      -H "Content-Type: application/json" \
      --data "$body" \
      "$BASE_URL$path" | print_response
    return
  fi

  curl --silent --show-error \
    -X "$method" \
    "$BASE_URL$path" | print_response
}

build_action_body() {
  local seat=""
  local kind=""
  local cards=""
  local resolution_id=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --seat)
        seat="$2"
        shift 2
        ;;
      --kind)
        kind="$2"
        shift 2
        ;;
      --cards)
        cards="$2"
        shift 2
        ;;
      --resolution-id|--resolution)
        resolution_id="$2"
        shift 2
        ;;
      *)
        printf 'unknown action argument: %s\n' "$1" >&2
        exit 1
        ;;
    esac
  done

  if [[ -z "$seat" || -z "$kind" ]]; then
    printf 'action requires --seat and --kind\n' >&2
    exit 1
  fi

  printf '{"seat":"%s","kind":"%s","cards":%s' \
    "$(json_escape "$seat")" \
    "$(json_escape "$kind")" \
    "$(json_array_from_csv "$cards")"

  if [[ -n "$resolution_id" ]]; then
    printf ',"resolutionId":"%s"' "$(json_escape "$resolution_id")"
  fi
  printf '}'
}

if [[ $# -eq 0 ]]; then
  usage
  exit 1
fi

if [[ "${1:-}" == "--base-url" ]]; then
  BASE_URL="$2"
  shift 2
fi

command="${1:-}"
shift || true

case "$command" in
  state)
    curl_json GET /api/game/state
    ;;
  reset)
    curl_json POST /api/game/reset
    ;;
  prompt)
    curl_json GET /api/game/agent/prompt
    ;;
  trace)
    curl_json GET /api/game/agent/trace
    ;;
  match|match-trace)
    curl_json GET /api/game/agent/match/trace
    ;;
  match-state)
    curl_json GET /api/game/agent/match/state
    ;;
  match-reset)
    curl_json POST /api/game/agent/match/reset
    ;;
  action)
    if [[ $# -gt 0 && "${1:0:1}" == "{" ]]; then
      curl_json POST /api/game/action "$1"
    else
      curl_json POST /api/game/action "$(build_action_body "$@")"
    fi
    ;;
  run-mock)
    model=""
    if [[ "${1:-}" == "--model" ]]; then
      model="$2"
    fi
    if [[ -n "$model" ]]; then
      curl_json POST /api/game/agent/run "{\"mode\":\"mock\",\"model\":\"$(json_escape "$model")\"}"
    else
      curl_json POST /api/game/agent/run '{"mode":"mock"}'
    fi
    ;;
  run-openrouter)
    model=""
    if [[ "${1:-}" == "--model" ]]; then
      model="$2"
    fi
    if [[ -n "$model" ]]; then
      curl_json POST /api/game/agent/run "{\"mode\":\"openrouter\",\"model\":\"$(json_escape "$model")\"}"
    else
      curl_json POST /api/game/agent/run '{"mode":"openrouter"}'
    fi
    ;;
  match-run)
    model=""
    max_steps=""
    fallback_mode=""
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --model)
          model="$2"
          shift 2
          ;;
        --max-steps)
          max_steps="$2"
          shift 2
          ;;
        --fallback-mode)
          fallback_mode="$2"
          shift 2
          ;;
        *)
          printf 'unknown match-run argument: %s\n' "$1" >&2
          exit 1
          ;;
      esac
    done
    body='{"mode":"openrouter","resetGame":true'
    if [[ -n "$model" ]]; then
      body="$body,\"model\":\"$(json_escape "$model")\""
    fi
    if [[ -n "$max_steps" ]]; then
      body="$body,\"maxSteps\":$max_steps"
    fi
    if [[ -n "$fallback_mode" ]]; then
      body="$body,\"fallbackMode\":\"$(json_escape "$fallback_mode")\""
    fi
    body="$body}"
    curl_json POST /api/game/agent/match/run "$body"
    ;;
  match-run-mock)
    max_steps=""
    if [[ "${1:-}" == "--max-steps" ]]; then
      max_steps="$2"
    fi
    body='{"mode":"mock","resetGame":true'
    if [[ -n "$max_steps" ]]; then
      body="$body,\"maxSteps\":$max_steps"
    fi
    body="$body}"
    curl_json POST /api/game/agent/match/run "$body"
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    printf 'unknown command: %s\n\n' "$command" >&2
    usage
    exit 1
    ;;
esac
