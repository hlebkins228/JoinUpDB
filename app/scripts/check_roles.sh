#!/usr/bin/env bash

set -euo pipefail

container="${POSTGRES_CONTAINER:-join_up_db}"
database="${POSTGRES_DB:-join_up}"
guest_password="${POSTGRES_GUEST_PASSWORD:-guest}"
user_password="${POSTGRES_APP_PASSWORD:-user}"
admin_password="${POSTGRES_ADMIN_PASSWORD:-admin}"
result_dir="$(cd -- "$(dirname -- "$0")" && pwd)"
failed=0

printf 'role,operation,expected,actual\n' > "$result_dir/permissions.csv"

check_access() {
    local role="$1"
    local password="$2"
    local operation="$3"
    local expected="$4"
    local query="$5"
    local actual

    if docker exec -e PGPASSWORD="$password" "$container" \
        psql -v ON_ERROR_STOP=1 -h 127.0.0.1 -U "$role" -d "$database" -c "$query" >/dev/null 2>&1; then
        actual="allowed"
    else
        actual="denied"
    fi

    printf '%s,%s,%s,%s\n' "$role" "$operation" "$expected" "$actual" >> "$result_dir/permissions.csv"
    if [[ "$actual" != "$expected" ]]; then
        failed=1
    fi
}

check_access joinup_guest "$guest_password" select_event allowed "select count(*) from event;"
check_access joinup_guest "$guest_password" update_event denied "update event set name = name where false;"
check_access joinup_guest "$guest_password" select_users denied "select count(*) from users;"
check_access joinup_user "$user_password" select_event allowed "select count(*) from event;"
check_access joinup_user "$user_password" update_event allowed "begin; update event set name = name where false; rollback;"
check_access joinup_admin "$admin_password" delete_event allowed "begin; delete from event where false; rollback;"

if [[ "$failed" -ne 0 ]]; then
    printf 'Фактические права отличаются от ожидаемых, см. %s/permissions.csv\n' "$result_dir" >&2
    exit 1
fi

printf 'Результаты сохранены в %s/permissions.csv\n' "$result_dir"
