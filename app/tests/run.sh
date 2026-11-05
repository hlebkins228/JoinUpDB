#!/usr/bin/env bash
set -euo pipefail
output=${REPORT_DIR:-/report}
mkdir -p "$output"
rm -rf "$output/results" "$output/allure-results" "$output/html" "$output/single"
mkdir -p "$output/results"
rm -f "$output/capture.status" "$output/replay.status" "$output/cleanup.status" "$output/e2e.pcap" "$output/traffic.log" "$output/report.html"
for stage in unit integration e2e; do echo skipped > "$output/$stage.status"; done
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
    if [[ -s "$output/e2e.pcap" ]]; then
        tcpdump -n -A -r "$output/e2e.pcap" > "$output/traffic.log" 2>&1
    fi
    if psql -v ON_ERROR_STOP=1 -f tests/reset.sql > "$output/cleanup.log" 2>&1 &&
        [[ $(psql -Atc "select (select count(*) from users) + (select count(*) from event) + (select count(*) from review) + (select count(*) from location) + (select count(*) from image)") == 0 ]]; then
        echo passed > "$output/cleanup.status"
    else
        echo failed > "$output/cleanup.status"
        code=1
    fi
    node tests/report.mjs "$output" || code=1
    if [[ -d "$output/history" ]]; then cp -a "$output/history" "$output/allure-results/history"; fi
    if allure generate "$output/allure-results" --clean -o "$output/html"; then
        rm -rf "$output/history"
        cp -a "$output/html/history" "$output/history"
    else
        code=1
    fi
    if allure generate "$output/allure-results" --single-file --clean -o "$output/single"; then
        cp "$output/single/index.html" "$output/report.html"
    else
        code=1
    fi
    chown -R "${TEST_UID:-1000}:${TEST_GID:-1000}" "$output"
    exit "$code"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
run_stage() {
    stage=$1
    shift
    echo running > "$output/$stage.status"
    if ALLURE_OUTPUT_PATH="$output/results" ALLURE_OUTPUT_FOLDER="$stage" \
        "$@" 2>&1 | tee "$output/$stage.log"; then
        echo passed > "$output/$stage.status"
    else
        echo failed > "$output/$stage.status"
        return 1
    fi
}
run_stage unit go test ./internal/service -count=1 -timeout=2m -v -coverpkg=./internal/service,./internal/repository/postgres -coverprofile="$output/unit.coverage.out"
for attempt in {1..60}; do
    if curl --noproxy '*' --silent --fail "$BACKEND_URL/api/v1/user/category" >/dev/null; then break; fi
    if [[ "$attempt" == 60 ]]; then
        echo failed > "$output/integration.status"
        echo 'Application did not become ready' > "$output/integration.log"
        exit 1
    fi
    sleep 1
done
run_stage integration go test -tags=integration ./internal/integration_test -count=1 -timeout=2m -v
psql -v ON_ERROR_STOP=1 -f tests/reset.sql

tcpdump -U -n -i any -w "$output/e2e.pcap" 'tcp port 8080' > "$output/tcpdump.log" 2>&1 &
capture=$!
sleep 1
if ! kill -0 "$capture" 2>/dev/null; then
    echo failed > "$output/capture.status"
    cp "$output/tcpdump.log" "$output/capture.log"
    exit 1
fi
run_stage e2e go test -tags=e2e ./internal/e2e_test -count=1 -timeout=2m -v
if ! bash tests/replay.sh 2>&1 | tee "$output/curl-replay.log"; then
    echo failed > "$output/replay.status"
    exit 1
fi
stop_capture
if [[ ! -s "$output/e2e.pcap" ]]; then
    echo failed > "$output/capture.status"
    exit 1
fi
