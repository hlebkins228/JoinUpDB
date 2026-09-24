import io
import json
import os
import random
from datetime import datetime, timedelta, timezone
from pathlib import Path

import psycopg
from faker import Faker
from PIL import Image, ImageDraw
from psycopg import sql


PASSWORD_HASH = "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi."
SEED_DATA = Path(__file__).with_name("seed_data.json")


def connection_params():
    return {
        "user": os.environ["POSTGRES_USER"],
        "password": os.environ["POSTGRES_PASSWORD"],
        "host": os.environ["POSTGRES_HOST"],
        "port": os.environ["POSTGRES_PORT"],
        "dbname": os.environ["POSTGRES_DB"],
    }


def generate_data(user_count, now):
    source = json.loads(SEED_DATA.read_text())
    fake = Faker("ru_RU")
    fake.seed_instance(42)
    rng = random.Random(42)
    data = {
        "image": ("id, name, data", []),
        "location": ("id, name, longitude, latitude, address", []),
        "users": ("id, name, age, login, password, created_at, city, telegram_login, avatar_id, role", []),
        "category": ("id, name, subcategory_id", []),
        "event": ("id, name, description, created_at, updated_at, event_date, telegram_chat_url, city, creator_id, location_id, image_id, deleted", []),
        "member": ("id, user_id, event_id, role", []),
        "event_category": ("event_id, category_id", []),
        "preset": ("id, name, category_id, creator_id", []),
        "subscribe": ("id, subscriber_id, user_id, subscribe_at", []),
        "review": ("id, event_id, author_id, rating, text, created_at", []),
    }
    cities = source["cities"]
    activities = source["activities"]
    users_by_city = {city["name"]: [] for city in cities}
    for user_id in range(1, user_count + 1):
        city = cities[(user_id - 1) % len(cities)]
        users_by_city[city["name"]].append(user_id)
        data["users"][1].append((
            user_id, fake.name(), rng.randint(18, 65), f"user_{user_id}", PASSWORD_HASH,
            now - timedelta(days=365), city["name"], f"joinup_{user_id}", user_id,
            "admin" if user_id == 1 else "user",
        ))
        picture = Image.new("RGB", (96, 96), tuple(rng.randint(80, 220) for _ in range(3)))
        ImageDraw.Draw(picture).text((12, 40), f"#{user_id}", fill="black")
        image_bytes = io.BytesIO()
        picture.save(image_bytes, format="PNG")
        data["image"][1].append((user_id, f"avatar_{user_id}.png", image_bytes.getvalue()))
    for category_id, activity in enumerate(activities, 1):
        data["category"][1].append((category_id, activity["category"], None))
    category_ids = {}
    for category_id, activity in enumerate(activities, 1):
        category_ids[category_id] = []
        for topic in activity["topics"]:
            for format_name in source["category_formats"]:
                child_id = len(data["category"][1]) + 1
                category_ids[category_id].append(child_id)
                data["category"][1].append((child_id, f"{topic}: {format_name}", category_id))
    events_by_city = {city["name"]: [] for city in cities}
    for user_id in range(1, user_count + 1):
        city = cities[(user_id - 1) % len(cities)]
        for _ in range(rng.randint(1, 3)):
            event_id = len(data["event"][1]) + 1
            category_id = rng.randrange(len(activities)) + 1
            activity = activities[category_id - 1]
            event_date = (now + timedelta(days=rng.randint(-90, 90))).replace(hour=18, minute=0, second=0)
            data["location"][1].append((
                event_id, activity["place"], city["longitude"] + rng.uniform(-0.06, 0.06),
                city["latitude"] + rng.uniform(-0.04, 0.04),
                f"{city['name']}, {fake.street_name()}, {rng.randint(1, 80)}",
            ))
            created_at = min(event_date - timedelta(days=rng.randint(7, 30)), now - timedelta(days=1))
            data["event"][1].append((
                event_id, activity["title"], activity["description"], created_at, created_at,
                event_date, f"https://t.me/joinup_event_{event_id}", city["name"], user_id,
                event_id, user_id, False,
            ))
            data["event_category"][1].append((event_id, category_id))
            data["event_category"][1].append((event_id, rng.choice(category_ids[category_id])))
            events_by_city[city["name"]].append((event_id, user_id, event_date))
        favourite = rng.randrange(len(activities)) + 1
        data["preset"][1].append((user_id, activities[favourite - 1]["title"], favourite, user_id))
    for user_id in range(1, user_count + 1):
        city = cities[(user_id - 1) % len(cities)]["name"]
        other_events = [event for event in events_by_city[city] if event[1] != user_id]
        for event_id, _, event_date in rng.sample(other_events, rng.randint(4, 9)):
            data["member"][1].append((len(data["member"][1]) + 1, user_id, event_id, "member"))
            if event_date < now and rng.random() < 0.8:
                rating = rng.choices([3, 4, 5], weights=[1, 3, 6])[0]
                text = rng.choice(source["reviews"][str(rating)])
                review_date = min(event_date + timedelta(hours=rng.randint(3, 48)), now)
                data["review"][1].append((len(data["review"][1]) + 1, event_id, user_id, rating, text, review_date))
        others = [other for other in users_by_city[city] if other != user_id]
        for target_id in rng.sample(others, rng.randint(2, 5)):
            data["subscribe"][1].append((len(data["subscribe"][1]) + 1, user_id, target_id, now - timedelta(days=rng.randint(1, 90))))
    return data


def main():
    now = datetime.now(timezone.utc).replace(tzinfo=None, microsecond=0)
    data = generate_data(5001, now)
    with psycopg.connect(**connection_params()) as connection:
        with connection.cursor() as cursor:
            cursor.execute("truncate event_deletion_log, review, subscribe, preset, event_category, category, member, event, users, location, image")
            for table, (columns, rows) in data.items():
                with cursor.copy(sql.SQL("copy {} ({}) from stdin").format(sql.Identifier(table), sql.SQL(columns))) as copy:
                    for row in rows:
                        copy.write_row(row)
                if table != "event_category":
                    cursor.execute("select setval(pg_get_serial_sequence(%s, 'id'), %s)", (table, len(rows)))
            cursor.execute("select setval('event_deletion_log_id_seq', 1, false)")
            cursor.execute("update event set deleted = true, updated_at = %s where id %% 20 = 0", (now,))
            for table in (*data, "event_deletion_log"):
                cursor.execute(sql.SQL("select count(*) from {}").format(sql.Identifier(table)))
                print(f"{table}: {cursor.fetchone()[0]}")
    print("Вход: user_1 (администратор), user_2 (пользователь); пароль: password")


if __name__ == "__main__":
    main()
