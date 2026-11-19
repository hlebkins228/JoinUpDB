#!/usr/bin/env bash
set -euo pipefail
url="$BACKEND_URL/api/v1"
request() { curl --noproxy '*' --fail-with-body --silent --show-error "$@"; }
request -H 'Content-Type: application/json' -d \
    '{"name":"Replay User","age":25,"login":"replay_user","password":"password","city":"Moscow"}' \
    "$url/users"
token=$(request -X POST -H 'login: replay_user' -H 'password: password' "$url/tokens" | jq -er .token)
event=$(request -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d \
    '{"name":"Replay meetup","city":"Moscow","event_time":"2100-01-01T12:00:00Z","location":{"name":"Park","latitude":55.75,"longitude":37.62,"address":"Park entrance"}}' \
    "$url/events" | jq -er .id)
request "$url/events/$event" | jq -e '.name == "Replay meetup" and .deleted == false'
request -X DELETE -H "Authorization: Bearer $token" "$url/events/$event"
request "$url/events/$event" | jq -e '.deleted == true'
