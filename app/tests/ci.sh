#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if ! timeout 10s docker info >/dev/null 2>&1; then
    echo 'Docker недоступен. Включите Docker Desktop / Docker Engine и повторите запуск.' >&2
    exit 1
fi
export COMPOSE_PROJECT_NAME="joinup-tests-$(id -u)-$$"
export TEST_REPORT_DIR="${TEST_REPORT_DIR:-$PWD/test-report/ci/$COMPOSE_PROJECT_NAME}"
export TEST_UID=$(id -u)
export TEST_GID=$(id -g)
mkdir -p "$TEST_REPORT_DIR"
compose() { docker compose -f tests/compose.yaml "$@"; }
cleanup() {
    code=$?
    trap - EXIT
    compose logs --no-color db app > "$TEST_REPORT_DIR/stand.log" 2>&1 || true
    compose down --volumes --remove-orphans || code=1
    exit "$code"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
compose build
compose up -d db app
compose run --rm tests
printf 'Report: %s/report.html\n' "$TEST_REPORT_DIR"
