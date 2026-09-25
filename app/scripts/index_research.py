import csv
import json
import math
import os
import platform
import random
import statistics
from datetime import datetime, timedelta, timezone
from pathlib import Path
from importlib.metadata import version as package_version

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
import psycopg

from seed import connection_params, generate_data


DIRECTORY = Path(__file__).resolve().parent
SIZES = (1000, 2000, 5000, 10000)
BLOCKS = 10
REPEATS = 50
WARMUP = 10
QUERY = """
select e.id, e.name, e.description,
       e.created_at at time zone 'utc', e.updated_at at time zone 'utc',
       e.event_date at time zone 'utc', e.telegram_chat_url, e.city,
       e.creator_id, e.image_id, l.name, l.longitude, l.latitude, l.address, e.deleted
from event e
join location l on e.location_id = l.id
where e.deleted = false
  and (%(name)s::text is null or e.name ilike '%%' || %(name)s || '%%')
  and (%(date_from)s::timestamp is null or e.event_date >= %(date_from)s)
  and (%(date_to)s::timestamp is null or e.event_date <= %(date_to)s)
  and (%(city)s::text is null or e.city = %(city)s)
  and (cardinality(%(categories)s::int[]) = 0 or e.id in (
      select ec.event_id from event_category ec
      where ec.category_id = any(%(categories)s::int[])
      group by ec.event_id
      having count(distinct ec.category_id) = cardinality(%(categories)s::int[])
  ))
order by e.event_date, e.id
"""
INDEX = "create index research_city_date_idx on event (city, event_date, id) where deleted = false"
EXPLAIN = "explain (analyze, buffers, timing off, format json) " + QUERY


def plot_chart(summaries):
    fig, ax = plt.subplots(figsize=(8, 5), layout="constrained")
    for variant, label, linestyle, marker in (
        ("without_index", "Без индекса", "-", "^"),
        ("with_index", "С индексом", "--", "s"),
    ):
        values = [s for s in summaries if s["variant"] == variant]
        ax.errorbar([s["event_count"] for s in values], [s["mean_ms"] for s in values],
                    yerr=[s["ci95_half_width_ms"] for s in values], color="black",
                    linestyle=linestyle, marker=marker, markerfacecolor="white", capsize=4, label=label)
    ax.set(xlabel="Количество мероприятий", ylabel="Среднее время выполнения запроса, мс",
           title="Зависимость времени выполнения запроса от количества записей в таблице", ylim=(0, None))
    ax.set_xticks(SIZES)
    ax.grid(alpha=0.3)
    ax.legend()
    fig.savefig(DIRECTORY / "index_chart.svg")
    plt.close(fig)


def main():
    reference_time = datetime.now(timezone.utc).replace(tzinfo=None, microsecond=0)
    data = generate_data(6000, reference_time)
    events = data["event"][1]
    random.Random(73).shuffle(events)
    params = {
        "name": None,
        "date_from": reference_time - timedelta(days=30),
        "date_to": reference_time + timedelta(days=30),
        "city": "Москва",
        "categories": [],
    }
    summaries = []
    samples = []
    plans = []
    with psycopg.connect(**connection_params(), autocommit=True, prepare_threshold=None) as connection:
        with connection.cursor() as cursor:
            cursor.execute("set timezone = 'UTC'")
            cursor.execute("set jit = off")
            cursor.execute("set max_parallel_workers_per_gather = 0")
            cursor.execute("select version()")
            version = cursor.fetchone()[0]
            cursor.execute("select name, setting, unit from pg_settings where name = any(%s)",
                           (["shared_buffers", "work_mem", "random_page_cost", "effective_cache_size",
                             "jit", "max_parallel_workers_per_gather", "block_size", "temp_buffers"],))
            settings = cursor.fetchall()
            cursor.execute("create temp table location (like public.location including constraints)")
            cursor.execute("alter table location add primary key (id)")
            columns, locations = data["location"]
            with cursor.copy(f"copy location ({columns}) from stdin") as copy:
                for row in locations:
                    copy.write_row(row)
            cursor.execute("analyze location")
            cursor.execute("create temp table event (like public.event including constraints)")
            cursor.execute("alter table event add primary key (id)")
            cursor.execute("create temp table event_category (event_id int, category_id int, primary key (event_id, category_id))")
            cursor.execute("select count(*) from location")
            location_count = cursor.fetchone()[0]
            for size in SIZES:
                cursor.execute("truncate event")
                columns = data["event"][0]
                with cursor.copy(f"copy event ({columns}) from stdin") as copy:
                    for row in events[:size]:
                        copy.write_row(row)
                cursor.execute("update event set deleted = true where id % 20 = 0")
                cursor.execute("analyze event")
                for block in range(BLOCKS):
                    variants = ("without_index", "with_index") if block % 2 == 0 else ("with_index", "without_index")
                    for variant in variants:
                        cursor.execute("drop index if exists pg_temp.research_city_date_idx")
                        if variant == "with_index":
                            cursor.execute(INDEX)
                        for _ in range(WARMUP):
                            cursor.execute(EXPLAIN, params)
                            cursor.fetchone()
                        for repeat in range(REPEATS):
                            cursor.execute(EXPLAIN, params)
                            plan = cursor.fetchone()[0][0]
                            samples.append({
                                "event_count": size, "variant": variant, "block": block + 1,
                                "repeat": repeat + 1, "execution_ms": plan["Execution Time"],
                                "planning_ms": plan["Planning Time"], "returned_rows": plan["Plan"]["Actual Rows"],
                            })
                            if block == 0 and repeat == 0:
                                plans.append({"event_count": size, "variant": variant, "explain": plan})
                for variant in ("without_index", "with_index"):
                    chosen = [s for s in samples if s["event_count"] == size and s["variant"] == variant]
                    times = [s["execution_ms"] for s in chosen]
                    block_means = [statistics.mean(s["execution_ms"] for s in chosen if s["block"] == b)
                                   for b in range(1, BLOCKS + 1)]
                    summaries.append({
                        "event_count": size, "variant": variant, "measurements": len(times),
                        "mean_ms": statistics.mean(times), "stddev_ms": statistics.stdev(times),
                        "ci95_half_width_ms": 2.262 * statistics.stdev(block_means) / math.sqrt(BLOCKS),
                        "returned_rows": chosen[0]["returned_rows"],
                    })
                print(f"Измерено: {size} мероприятий, по {BLOCKS * REPEATS} замеров на вариант", flush=True)
    with (DIRECTORY / "index_samples.csv").open("w", newline="") as output:
        writer = csv.DictWriter(output, fieldnames=list(samples[0]))
        writer.writeheader()
        writer.writerows(samples)
    report = {
        "reference_time_utc": reference_time.isoformat(), "postgresql": version,
        "client_platform": platform.platform(), "client_cpu_count": os.cpu_count(),
        "packages": {name: package_version(name) for name in ("psycopg", "Faker", "matplotlib", "Pillow")},
        "postgresql_settings": settings, "location_count_fixed": location_count,
        "method": {
            "axis": "Количество мероприятий; набор фильтров и число локаций фиксированы",
            "scope": "Основной SQL поиска из EventRepo.SearchEvents, без последующих загрузок связей и HTTP",
            "timing": "Серверный Execution Time из EXPLAIN (ANALYZE, BUFFERS, TIMING OFF); планирование отдельно в CSV",
            "blocks": BLOCKS, "measurements_per_block": REPEATS,
            "warmup_per_variant_per_block": WARMUP,
            "order": "Порядок с индексом/без индекса чередуется между блоками",
            "cache": "Прогретый кеш; сброс кеша ОС и PostgreSQL не выполняется",
            "dataset": "Фиксированная генерация, перемешивание seed=73; вложенные префиксы одного набора; временные таблицы",
            "outliers": "Все замеры включены в арифметическое среднее, выбросы не удаляются",
            "ci": "Оценка 95% ДИ по 10 средним блоков, t(9)=2.262; блоки могут быть коррелированы",
            "limits": "Временные таблицы используют local buffers; это сравнение SQL на контролируемых данных, не нагрузочный тест сайта",
        },
        "query": QUERY, "parameters": params, "summary": summaries, "plans": plans,
    }
    (DIRECTORY / "index_results.json").write_text(json.dumps(report, ensure_ascii=False, indent=2, default=str) + "\n")
    plot_chart(summaries)
    for result in summaries:
        print(f"{result['event_count']}: {result['variant']} = {result['mean_ms']:.4f} мс")


if __name__ == "__main__":
    main()
