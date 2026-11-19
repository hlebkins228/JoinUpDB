#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if ! timeout 10s docker info >/dev/null 2>&1; then
    echo 'Docker not found' >&2
    exit 1
fi
stage=${1:-all}
export COMPOSE_PROJECT_NAME="joinup-tests-$(id -u)-$$"
export TEST_REPORT_DIR="${TEST_REPORT_DIR:-$PWD/test-report/ci/$COMPOSE_PROJECT_NAME}"
export TEST_UID=$(id -u)
export TEST_GID=$(id -g)
mkdir -p "$TEST_REPORT_DIR"
compose_files=(-f tests/compose.yaml)
windows=false
windows_stand=false
if [[ "$stage" == e2e && "${E2E_PLATFORM:-docker}" == windows ]]; then
    windows=true
    compose_files+=(-f tests/compose.windows.yaml)
fi
compose() { docker compose "${compose_files[@]}" "$@"; }
if [[ "$stage" == prepare || ( "${GITHUB_ACTIONS:-}" == true && "$stage" != report ) ]]; then
    compose pull --policy always db
    compose build app tests
    if [[ "$stage" == prepare ]]; then
        echo 'Test images prepared. Test runs can now work offline.'
        exit 0
    fi
fi
images=(joinup-test-runner:local)
if [[ "$stage" != report ]]; then images+=(joinup-test-app:local postgres:16.6-bookworm); fi
for image in "${images[@]}"; do
    if ! docker image inspect "$image" >/dev/null 2>&1; then
        echo "Missing local image: $image. Run make test-prepare with internet access or import images with docker load." >&2
        exit 1
    fi
done
if [[ "$stage" == report ]]; then
    docker run --rm --pull=never --network none --user "$TEST_UID:$TEST_GID" \
        -v "$TEST_REPORT_DIR:/report" --entrypoint node joinup-test-runner:local \
        tests/report.mjs /report --generate
    exit 0
fi
cleanup() {
    code=$?
    trap - EXIT
    if $windows; then
        if [[ $(cat "$TEST_REPORT_DIR/e2e.status") == running ]]; then
            echo failed > "$TEST_REPORT_DIR/e2e.status"
        fi
        if $windows_stand; then
            if compose exec -T db psql -U postgres -d joinup_test -v ON_ERROR_STOP=1 \
                < tests/reset.sql > "$TEST_REPORT_DIR/cleanup.log" 2>&1; then
                echo passed > "$TEST_REPORT_DIR/cleanup.status"
            else
                echo failed > "$TEST_REPORT_DIR/cleanup.status"
                code=1
            fi
        fi
    fi
    compose logs --no-color db app > "$TEST_REPORT_DIR/stand.log" 2>&1 || true
    compose down --volumes --remove-orphans || code=1
    exit "$code"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
if $windows; then
    rm -rf "$TEST_REPORT_DIR/results/e2e"
    rm -f "$TEST_REPORT_DIR/"{capture,replay,cleanup}.status "$TEST_REPORT_DIR/e2e.pcap" \
        "$TEST_REPORT_DIR/traffic.log" "$TEST_REPORT_DIR/report.html"
    mkdir -p "$TEST_REPORT_DIR/results/e2e"
    echo running > "$TEST_REPORT_DIR/e2e.status"
    docker run --rm --pull=never --network none --user "$TEST_UID:$TEST_GID" \
        -v "$TEST_REPORT_DIR:/report" --entrypoint cp joinup-test-runner:local \
        /app/e2e.test.exe /report/e2e.test.exe \
        2>&1 | tee "$TEST_REPORT_DIR/e2e.log"
fi
if [[ "$stage" == integration ]]; then
    compose up -d --wait --no-build --pull never db
else
    compose up -d --no-build --pull never db app
fi
if $windows; then
    windows_stand=true
    backend="http://127.0.0.1:${E2E_PORT:-18080}"
    for attempt in {1..60}; do
        if curl.exe --noproxy '*' --silent --fail --connect-timeout 1 --max-time 2 \
            "$backend/api/v1/categories" >/dev/null; then break; fi
        if [[ "$attempt" == 60 ]]; then
            echo "Windows cannot reach $backend" | tee -a "$TEST_REPORT_DIR/e2e.log"
            exit 1
        fi
        sleep 1
    done
    printf 'Windows e2e: %s (Wireshark: Adapter for loopback traffic capture)\n' "$backend"
    if WSLENV="${WSLENV:+$WSLENV:}BACKEND_URL:ALLURE_OUTPUT_PATH/p:ALLURE_OUTPUT_FOLDER" \
        BACKEND_URL="$backend" ALLURE_OUTPUT_PATH="$TEST_REPORT_DIR/results" ALLURE_OUTPUT_FOLDER=e2e \
        "$TEST_REPORT_DIR/e2e.test.exe" -test.v -test.count=1 -test.timeout=2m \
        2>&1 | tee -a "$TEST_REPORT_DIR/e2e.log"; then
        echo passed > "$TEST_REPORT_DIR/e2e.status"
    else
        echo failed > "$TEST_REPORT_DIR/e2e.status"
        exit 1
    fi
    if ! compose run --rm --no-deps --pull never --entrypoint bash tests tests/replay.sh \
        2>&1 | tee "$TEST_REPORT_DIR/replay.log"; then
        echo failed > "$TEST_REPORT_DIR/replay.status"
        exit 1
    fi
else
    compose run --rm --no-deps --pull never tests "$stage"
fi
printf 'Results: %s\n' "$TEST_REPORT_DIR"
