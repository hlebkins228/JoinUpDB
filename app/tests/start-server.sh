#!/usr/bin/env bash
set -euo pipefail
secret=$(cat /proc/sys/kernel/random/uuid)
sed "s/TEST_JWT_SECRET/$secret/" server.yaml > config.yaml
./migrate up
exec ./server
