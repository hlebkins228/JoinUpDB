import os
import time
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

import psycopg


API_URL = "https://api.dicebear.com/10.x"
REQUEST_INTERVAL = 0.1
BATCH_SIZE = 100
PNG_HEADER = b"\x89PNG\r\n\x1a\n"


def connection_params():
    return {
        "user": os.environ["POSTGRES_USER"],
        "password": os.environ["POSTGRES_PASSWORD"],
        "host": os.environ["POSTGRES_HOST"],
        "port": os.environ["POSTGRES_PORT"],
        "dbname": os.environ["POSTGRES_DB"],
    }


def download(style, seed, last_request_at):
    url = f"{API_URL}/{style}/png?{urlencode({'seed': seed, 'size': 256})}"
    for attempt in range(5):
        remaining = REQUEST_INTERVAL - (time.monotonic() - last_request_at)
        if remaining > 0:
            time.sleep(remaining)
        last_request_at = time.monotonic()
        try:
            with urlopen(Request(url, headers={"User-Agent": "JoinUpDB educational seed"}), timeout=30) as response:
                data = response.read()
            if not data.startswith(PNG_HEADER):
                raise ValueError(f"DiceBear did not return PNG: {url}")
            return data, last_request_at
        except HTTPError as error:
            if attempt == 4 or error.code not in (429, 500, 502, 503, 504):
                raise
            delay = error.headers.get("Retry-After") if error.code == 429 else None
            time.sleep(max(10, int(delay)) if delay and delay.isdecimal() else 10)
        except URLError:
            if attempt == 4:
                raise
            time.sleep(10)


def replace_avatar(cursor, user_id, data):
    cursor.execute("""
        select u.avatar_id, i.name
        from users u join image i on i.id = u.avatar_id
        where u.id = %s
        for update of u
    """, (user_id,))
    image_id, image_name = cursor.fetchone()
    name = f"dicebear_avatar_{user_id}.png"
    if image_name == name:
        return False
    cursor.execute("""
        select exists(select 1 from users where avatar_id = %s and id <> %s)
            or exists(select 1 from event where image_id = %s)
    """, (image_id, user_id, image_id))
    if cursor.fetchone()[0]:
        cursor.execute("insert into image (name, data) values (%s, %s) returning id", (name, data))
        cursor.execute("update users set avatar_id = %s where id = %s", (cursor.fetchone()[0], user_id))
    else:
        cursor.execute("update image set name = %s, data = %s where id = %s", (name, data, image_id))
    return True


def replace_event_image(cursor, event_id, data):
    cursor.execute("""
        select e.image_id, i.name
        from event e join image i on i.id = e.image_id
        where e.id = %s
        for update of e
    """, (event_id,))
    image_id, image_name = cursor.fetchone()
    name = f"dicebear_event_{event_id}.png"
    if image_name == name:
        return False
    cursor.execute("""
        select exists(select 1 from event where image_id = %s and id <> %s)
            or exists(select 1 from users where avatar_id = %s)
    """, (image_id, event_id, image_id))
    if cursor.fetchone()[0]:
        cursor.execute("insert into image (name, data) values (%s, %s) returning id", (name, data))
        cursor.execute("update event set image_id = %s where id = %s", (cursor.fetchone()[0], event_id))
    else:
        cursor.execute("update image set name = %s, data = %s where id = %s", (name, data, image_id))
    return True


def main():
    with psycopg.connect(**connection_params(), autocommit=True) as connection:
        with connection.cursor() as cursor:
            cursor.execute("""
                select u.id, i.name
                from users u join image i on i.id = u.avatar_id
                order by u.id
            """)
            users = cursor.fetchall()
            cursor.execute("""
                select e.id, i.name
                from event e join image i on i.id = e.image_id
                order by e.id
            """)
            events = cursor.fetchall()

        pending = [
            ("avatar", user_id)
            for user_id, image_name in users
            if image_name != f"dicebear_avatar_{user_id}.png"
        ] + [
            ("event", event_id)
            for event_id, image_name in events
            if image_name != f"dicebear_event_{event_id}.png"
        ]
        total = len(pending)
        print(f"К загрузке: {total} изображений, пачками по {BATCH_SIZE}", flush=True)
        avatars_replaced = 0
        events_replaced = 0
        last_request_at = 0.0
        for start in range(0, total, BATCH_SIZE):
            batch = []
            for kind, item_id in pending[start:start + BATCH_SIZE]:
                style = "lorelei" if kind == "avatar" else "landscape"
                data, last_request_at = download(style, f"joinup-{kind}-{item_id}", last_request_at)
                batch.append((kind, item_id, data))
            with connection.transaction():
                with connection.cursor() as cursor:
                    for kind, item_id, data in batch:
                        if kind == "avatar":
                            avatars_replaced += replace_avatar(cursor, item_id, data)
                        else:
                            events_replaced += replace_event_image(cursor, item_id, data)
            print(
                f"Скачано и сохранено {min(start + BATCH_SIZE, total)}/{total}; "
                f"аватаров: {avatars_replaced}, обложек: {events_replaced}",
                flush=True,
            )

    print(f"Заменено аватаров: {avatars_replaced}; обложек: {events_replaced}")


if __name__ == "__main__":
    main()
