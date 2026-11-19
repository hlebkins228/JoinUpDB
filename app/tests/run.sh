#!/usr/bin/env bash
set -euo pipefail
stage=${1:-all}
output=${REPORT_DIR:-/report}
stages=("$stage")
if [[ "$stage" == all ]]; then stages=(unit integration e2e); fi
mkdir -p "$output/results"
for selected in "${stages[@]}"; do
    rm -rf "$output/results/$selected"
    rm -f "$output/$selected.log" "$output/$selected.coverage.out"
    echo skipped > "$output/$selected.status"
done
if [[ "$stage" != unit ]]; then
    rm -f "$output/cleanup.status"
fi
if [[ "$stage" == all || "$stage" == e2e ]]; then
    rm -f "$output/"{capture,replay}.status "$output/e2e.pcap" "$output/traffic.log"
fi
rm -f "$output/report.html"
capture=''
stop_capture() {
    if [[ -n "$capture" ]]; then
        kill -INT "$capture" 2>/dev/null || true
        wait "$capture" || true
        capture=''
    fi
}
finish() {
    code=$?
    trap - EXIT
    set +e
    stop_capture
    if [[ "$stage" == all || "$stage" == e2e ]] && [[ -s "$output/e2e.pcap" ]]; then
        tcpdump -n -A -r "$output/e2e.pcap" > "$output/traffic.log" 2>&1
    fi
    if [[ "$stage" != unit ]]; then
        if psql -v ON_ERROR_STOP=1 -f tests/reset.sql > "$output/cleanup.log" 2>&1 &&
            [[ $(psql -Atc "select (select count(*) from users) + (select count(*) from event) + (select count(*) from review) + (select count(*) from location) + (select count(*) from image)") == 0 ]]; then
            echo passed > "$output/cleanup.status"
        else
            echo failed > "$output/cleanup.status"
            code=1
        fi
    fi
    if [[ "$stage" == all ]]; then node tests/report.mjs "$output" --generate || code=1; fi
    if [[ -n "${TEST_UID:-}" ]]; then chown -R "$TEST_UID:$TEST_GID" "$output"; fi
    exit "$code"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
run_stage() {
    selected=$1
    shift
    echo running > "$output/$selected.status"
    if ALLURE_OUTPUT_PATH="$output/results" ALLURE_OUTPUT_FOLDER="$selected" \
        "$@" 2>&1 | tee "$output/$selected.log"; then
        echo passed > "$output/$selected.status"
    else
        echo failed > "$output/$selected.status"
        return 1
    fi
}
if [[ "$stage" == all || "$stage" == unit ]]; then
    run_stage unit go test ./internal/service -count=1 -timeout=2m -v -coverpkg=./internal/service,./internal/repository/postgres -coverprofile="$output/unit.coverage.out"
fi
if [[ "$stage" == integration ]]; then
    echo running > "$output/integration.status"
    POSTGRES_HOST="$PGHOST" POSTGRES_PORT="$PGPORT" POSTGRES_USER="$PGUSER" \
        POSTGRES_PASSWORD="$PGPASSWORD" POSTGRES_DB="$PGDATABASE" ./migrate up 2>&1 | tee "$output/integration.log"
fi
if [[ "$stage" == all || "$stage" == e2e ]]; then
    for attempt in {1..60}; do
        if curl --noproxy '*' --silent --fail "$BACKEND_URL/api/v1/categories" >/dev/null; then break; fi
        if [[ "$attempt" == 60 ]]; then
            selected=e2e
            if [[ "$stage" == all ]]; then selected=integration; fi
            echo failed > "$output/$selected.status"
            echo 'Application did not become ready' > "$output/$selected.log"
            exit 1
        fi
        sleep 1
    done
fi
if [[ "$stage" == all || "$stage" == integration ]]; then
    run_stage integration go test -tags=integration ./internal/integration_test -count=1 -timeout=2m -v
fi
if [[ "$stage" == all || "$stage" == e2e ]]; then
    psql -v ON_ERROR_STOP=1 -f tests/reset.sql
    tcpdump -U -n -i any -w "$output/e2e.pcap" 'tcp port 8080' > "$output/capture.log" 2>&1 &
    capture=$!
    sleep 1
    if ! kill -0 "$capture" 2>/dev/null; then
        echo failed > "$output/capture.status"
        exit 1
    fi
    run_stage e2e go test -tags=e2e ./internal/e2e_test -count=1 -timeout=2m -v
    if ! bash tests/replay.sh 2>&1 | tee "$output/replay.log"; then
        echo failed > "$output/replay.status"
        exit 1
    fi
    stop_capture
    if [[ ! -s "$output/e2e.pcap" ]]; then
        echo failed > "$output/capture.status"
        exit 1
    fi
fi
