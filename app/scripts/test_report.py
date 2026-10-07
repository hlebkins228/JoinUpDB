import argparse
from collections import Counter, defaultdict
from html import escape
import json
from pathlib import Path
import subprocess


def prepare(results_dir, report_dir):
    results_dir.mkdir(parents=True, exist_ok=True)
    report_dir.mkdir(parents=True, exist_ok=True)
    for pattern in ("*-result.json", "*-container.json"):
        for path in results_dir.glob(pattern):
            path.unlink()
    for name in ("coverage.out", "coverage.html", "coverage.txt", "index.html"):
        (report_dir / name).unlink(missing_ok=True)


def build(results_dir, report_dir):
    profile = report_dir / "coverage.out"
    subprocess.run(
        ["go", "tool", "cover", f"-html={profile}", f"-o={report_dir / 'coverage.html'}"],
        check=True,
    )
    coverage_text = subprocess.check_output(
        ["go", "tool", "cover", f"-func={profile}"], text=True
    )
    (report_dir / "coverage.txt").write_text(coverage_text)
    total = coverage_text.splitlines()[-1].split()[-1]

    blocks = {}
    for line in profile.read_text().splitlines()[1:]:
        position, statements, count = line.split()
        previous = blocks.get(position, (int(statements), 0))
        blocks[position] = (int(statements), previous[1] + int(count))
    files = defaultdict(lambda: [0, 0])
    for position, (statements, count) in blocks.items():
        filename = position.rsplit(":", 1)[0]
        files[filename][0] += statements
        if count:
            files[filename][1] += statements
    file_rows = "".join(
        f"<tr><td>{escape(name)}</td><td>{100 * covered / statements:.1f}%</td></tr>"
        for name, (statements, covered) in sorted(files.items())
        if statements
    )

    results = [
        json.loads(path.read_text()) for path in results_dir.glob("*-result.json")
    ]
    statuses = Counter(result["status"] for result in results)
    test_rows = []
    for result in sorted(results, key=lambda item: item["fullName"]):
        suite = next(
            label["value"] for label in result["labels"] if label["name"] == "suite"
        )
        duration = (result["stop"] - result["start"]) / 1000
        status = escape(result["status"])
        test_rows.append(
            f"<tr><td>{escape(suite)}</td><td>{escape(result['name'])}</td>"
            f"<td class='{status}'>{status}</td><td>{duration:.3f} с</td></tr>"
        )

    page = """<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>JoinUp — отчёт тестирования</title>
<style>
body { font: 16px/1.6 system-ui, sans-serif; color: #17202b; background: #f3f5f7; margin: 0; }
main { max-width: 1180px; margin: 40px auto; padding: 0 24px 40px; }
h1 { font-size: 30px; margin-bottom: 4px; }
h2 { font-size: 20px; margin-top: 0; }
p { margin-top: 0; }
section, .card { background: white; border: 1px solid #d6dde4; border-radius: 10px; padding: 24px; }
section { margin-top: 24px; overflow-x: auto; }
.cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin: 24px 0; }
.value { display: block; font-size: 32px; font-weight: 700; }
.links { display: flex; flex-wrap: wrap; gap: 12px; }
a { color: #164d8c; }
.links a { padding: 10px 16px; background: white; border: 1px solid #9eaebb; border-radius: 6px; text-decoration: none; }
table { border-collapse: collapse; width: 100%; font-size: 14px; }
th, td { text-align: left; padding: 12px 10px; border-bottom: 1px solid #d6dde4; }
th { background: #f3f5f7; }
.passed { font-weight: 700; }
.failed, .broken { font-weight: 700; text-decoration: underline; }
.muted { color: #526273; font-size: 14px; }
</style>
</head>
<body><main>
<h1>JoinUp · отчёт тестирования</h1>
<p class="muted">Лабораторная № 1 · отзывы · соединение и транзакции в памяти</p>
<div class="cards">
<div class="card">Сценариев<span class="value">{tests}</span></div>
<div class="card">Успешно<span class="value">{passed}</span></div>
<div class="card">Ошибок<span class="value">{failed}</span></div>
<div class="card">Покрытие пакетов<span class="value">{coverage}</span></div>
</div>
<nav class="links">
<a href="allure/">Allure: тесты, фикстуры и время</a>
<a href="coverage.html">Покрытие: исходный код</a>
<a href="coverage.txt">Покрытие по функциям</a>
</nav>
<section><h2>Сценарии</h2>
<table><thead><tr><th>Набор</th><th>Тест</th><th>Статус</th><th>Время</th></tr></thead>
<tbody>{test_rows}</tbody></table></section>
<section><h2>Покрытие по файлам</h2>
<p class="muted">Измеряется доля выполненных операторов в пакетах internal/service и internal/repository/postgres
только девятью сценариями ЛР1. Остальные функции этих пакетов также входят в знаменатель.</p>
<table><thead><tr><th>Файл</th><th>Покрытие</th></tr></thead><tbody>{file_rows}</tbody></table>
</section></main></body></html>"""
    page = page.replace("{tests}", str(len(results)))
    page = page.replace("{passed}", str(statuses["passed"]))
    page = page.replace("{failed}", str(statuses["failed"] + statuses["broken"]))
    page = page.replace("{coverage}", escape(total))
    page = page.replace("{test_rows}", "".join(test_rows))
    page = page.replace("{file_rows}", file_rows)
    (report_dir / "index.html").write_text(page)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("action", choices=("prepare", "build"))
    parser.add_argument("--results-dir", type=Path, default=Path("allure-results"))
    parser.add_argument("--report-dir", type=Path, default=Path("test-report"))
    args = parser.parse_args()
    if args.action == "prepare":
        prepare(args.results_dir, args.report_dir)
    else:
        build(args.results_dir, args.report_dir)


if __name__ == "__main__":
    main()
