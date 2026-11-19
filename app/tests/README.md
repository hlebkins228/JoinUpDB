# Запуск тестов

Команды выполняются из `app`:

| Команда | Что делает | Зависимости на компьютере |
| --- | --- | --- |
| `make test` | 10 unit-сценариев с моками и сбором покрытия | Go и локальный кеш модулей |
| `make test-integration` | 3 интеграционных теста с отдельным PostgreSQL | Docker и Compose |
| `make test-e2e` | 1 HTTP e2e, повтор через curl и запись трафика | Docker и Compose |
| `make allure-report` | Собирает общий статический `test-report/report.html` | Docker и подготовленный runner |

## Подготовка и работа без интернета

Один раз при доступном интернете выполни:

```sh
make test-prepare
```

Команда скачает PostgreSQL и соберёт `joinup-test-app:local` и
`joinup-test-runner:local`. В runner находятся Go-модули, Allure, Java,
curl, jq, tcpdump и заранее собранный Windows e2e-тест.
Подготовка не запускает тесты.

После этого локальные `make test-integration`, `make test-e2e`,
`make test-e2e E2E_PLATFORM=windows`, `make test-ci` и `make allure-report`
работают с готовыми образами: сборка и скачивание при запуске отключены.
При отсутствии образа команда завершится с подсказкой, не обращаясь в registry.
Генерация локального отчёта выполняется в контейнере без сети.

Образы содержат снимок исходников. **После изменения сервера, тестов или
скриптов снова выполни `make test-prepare`**, чтобы запускать новую версию.
Кеш сборки сам по себе не гарантирует отсутствие обращения в registry;
поэтому для демонстрации офлайн используется именно запуск готовых образов.

Для переноса на другой компьютер с той же архитектурой:

```sh
docker save -o test-report/test-images.tar joinup-test-app:local joinup-test-runner:local postgres:16.6-bookworm
# На компьютере без интернета, после копирования архива:
docker load -i test-report/test-images.tar
```

Отдельная команда `make test` использует Go на хосте, без скачивания модулей
и toolchain. Для неё заранее нужны установленный Go подходящей версии
и заполненный кеш модулей (`go mod download` при доступном интернете).
Для полного прогона, включая unit, без зависимости от Go на хосте используй
`make test-ci` с подготовленными Docker-образами.

В GitHub Actions подготовка образов выполняется автоматически перед прогоном.
Аварийное формирование отчёта в CI, когда runner-образ собрать не удалось,
сохраняет установку Allure через npm на хосте.

Интеграционные тесты проверяют создание/чтение отзыва, отказ при дубликате
и отказ для будущего события. E2E: регистрация → вход → создание → чтение →
удаление события → проверка флага удаления.

Каждая группа запускается независимо. Интеграционным тестам автоматически
поднимается PostgreSQL и применяются миграции; e2e дополнительно запускает
сервер. После завершения или ошибки стенд удаляется. Рабочая БД не используется.
Образы должны быть заранее подготовлены через `make test-prepare` или импортированы.

Результаты групп сохраняются в `test-report/results`. Повторный запуск
заменяет результаты только своей группы. `make allure-report` не запускает
тесты: он объединяет последние результаты, а ещё не запущенные сценарии
отмечает skipped. После нового запуска отчёт нужно сформировать заново.
Открой `test-report/report.html` напрямую в браузере — сервер не нужен.

Для покрытия после `make test`:

```sh
go tool cover -func=test-report/unit.coverage.out
make coverage
```

HTML покрытия — `test-report/coverage.html`. Allure показывает результаты,
шаги и историю тестов; покрытие кода находится в отдельном HTML.

## Изоляция

Каждый Docker-запуск использует уникальное имя Compose project, собственную
сеть и БД `joinup_test` в tmpfs. В обычном режиме порты наружу не публикуются;
Windows-режим открывает HTTP-порт только на `127.0.0.1`. Фикстуры задают
фиксированные данные; таблицы очищаются до и после интеграционных тестов,
перед e2e и при завершении runner. Сервер получает новый JWT-секрет на запуск.

Для параллельных запусков задавай разные каталоги результатов, например
`make test-e2e TEST_REPORT_DIR=/tmp/joinup-e2e-1`. Для объединения результатов
и сохранения истории используй один каталог при последовательных запусках
и передавай его же в `make allure-report TEST_REPORT_DIR=...`.

## Полный прогон и CI

`make test-ci` оставлен для полного прогона по требованиям ЛР2:
unit → integration → e2e. После ошибки последующие этапы пропускаются,
отчёт формируется автоматически. Нужен только Docker с Compose.
Каталог отчёта — `test-report/ci/<имя проекта>`, путь выводится в терминале.
Его можно переопределить через `TEST_REPORT_DIR`.

GitHub Actions использует тот же полный прогон. Статический `report.html`
доступен в artifact `test-report-…`, логи и pcap — в `test-diagnostics-…`.
История Allure сохраняется в cache отдельно для каждой ветки. При ошибке
сборки Docker отчёт формируется на runner через ту же цель `allure-report`.

Для Pages выбери **Settings → Pages → Source: GitHub Actions**.
Публикация выполняется только из основной ветки репозитория; из остальных
веток и pull request отчёт доступен как artifact. URL показывает job `pages`.

## Зачем оставшиеся скрипты

- `ci.sh` — проверяет Docker, поднимает нужные контейнеры и удаляет стенд.
- `run.sh` — запускает выбранную группу или весь pipeline, сохраняет статусы и очищает БД.
- `report.mjs` — объединяет Allure-результаты, добавляет skipped/ошибки запуска, сохраняет историю и формирует HTML. Генерация общая для локального запуска и CI.
- `replay.sh` — повторяет HTTP-сценарий через curl, как требует ЛР2.

`compose.yaml` и `Dockerfile` описывают стенд, `server.yaml` — тестовую
конфигурацию сервера, `fixtures.sql` и `reset.sql` — данные и очистку.
Отдельный `start-server.sh` удалён: запуск сервера задан в Dockerfile.

После e2e в обычном Docker-режиме в каталоге результатов доступны `e2e.pcap`, `traffic.log`,
`replay.log`, `capture.log`, `stand.log` и `cleanup.log`.

## Живой захват в Windows Wireshark

Команда запускается из WSL, но сам e2e выполняется Windows-процессом.
Заранее собранный Go-тест копируется из runner в `test-report/e2e.test.exe`, сервер
и PostgreSQL остаются в Docker. HTTP-запросы идут через Windows loopback
на `http://127.0.0.1:18080`, опубликованный Docker Desktop.

Нужны подготовленные образы, Docker Desktop с интеграцией WSL, работающий
запуск `.exe` из WSL и доступный `curl.exe` из Windows. Go на хосте не нужен.
Для захвата нужны Windows Wireshark и Npcap с поддержкой loopback.

1. Открой Wireshark обычным способом в Windows.
2. Выбери **Adapter for loopback traffic capture** (`NPF_Loopback`).
3. В поле capture filter введи `tcp port 18080` и нажми Start.
4. В WSL из `app` выполни:

   ```sh
   make test-e2e E2E_PLATFORM=windows
   ```

5. В верхней строке display filter введи:

   ```text
   http.request && http.user_agent contains "Go-http-client"
   ```

   Для запросов вместе с ответами используй `http`. Если HTTP не распознан,
   выбери пакет порта 18080 и через **Analyze → Decode As…** назначь HTTP.

6. После завершения нажми Stop и сохрани захват через **File → Save As…**.

Авторизация выполняется отдельным `POST /api/v1/tokens` с заголовками
`Login: e2e_user` и `Password: password`, без JSON-тела. Сервер проверяет
эти данные и возвращает токен; последующие запросы отправляют его в
`Authorization: Bearer ...`.
Чтобы найти запрос авторизации, используй display filter
`http.request.method == "POST" && http.request.uri == "/api/v1/tokens"`.
В выбранном пакете раскрой **Hypertext Transfer Protocol** для просмотра
заголовков. Через **Follow → TCP Stream** видны и запрос, и ответ с токеном.
Захват нужно включить до запуска команды тестирования.

Перед e2e выполняются запросы готовности через Windows `curl.exe`; фильтр
по `Go-http-client` оставляет только запросы самого теста. После теста
curl-повтор выполняется внутри Docker, поэтому на Windows loopback не виден.
В этом режиме `e2e.pcap` автоматически не создаётся: захват делает Wireshark.
Старые pcap и результаты e2e удаляются перед запуском.

Allure-результаты записываются в общий каталог `test-report/results/e2e`.
Для отчёта после запуска выполни `make allure-report`. После завершения,
ошибки или Ctrl+C скрипт очищает и удаляет тестовый стенд.

Если порт занят, укажи другой и измени фильтр захвата соответственно:

```sh
make test-e2e E2E_PLATFORM=windows E2E_PORT=18081
```

Для параллельных Windows-запусков нужны разные `E2E_PORT` и `TEST_REPORT_DIR`.
Обычные `make test-e2e` и `make test-ci` по-прежнему выполняют тесты в Docker.

Справка: [захват loopback в Wireshark](https://wiki.wireshark.org/CaptureSetup/Loopback).

## Просмотр сохранённого захвата Docker-режима

В WSL из `app` выполни:

```sh
make test-e2e
cd test-report
explorer.exe .
```

Откроется папка результатов в Проводнике Windows. Открой `e2e.pcap`
через **File → Open** в Wireshark. Файл записывается tcpdump внутри
контейнера runner, поэтому выбирать сетевой адаптер Windows не нужно.

В строке display filter введи:

```text
http.request && http.user_agent contains "Go-http-client"
```

Так будут видны запросы Go-теста, без следующего за ним curl-повтора.
Ожидаются шесть запросов:

1. `POST /api/v1/users` — регистрация `e2e_user`.
2. `POST /api/v1/tokens` — получение токена по заголовкам `login` и `password`.
3. `POST /api/v1/events` — создание `E2E meetup`.
4. `GET /api/v1/events/<id>` — чтение.
5. `DELETE /api/v1/events/<id>` — удаление.
6. `GET /api/v1/events/<id>` — проверка `deleted: true`.

Для запросов и ответов вместе используй `http`. Выбери запрос, нажми
**Follow → TCP Stream** в контекстном меню: видны HTTP-метод, путь,
JSON-тело и ответы сервера. Для демонстрации достаточно списка шести
запросов и ответа последнего GET с `deleted: true`.
Если HTTP не распознан, выбери TCP-пакет порта 8080 и через
**Analyze → Decode As…** назначь этому порту протокол HTTP.

Справка: [файлы WSL из Windows](https://learn.microsoft.com/en-us/windows/wsl/filesystems),
[просмотр потоков Wireshark](https://www.wireshark.org/docs/wsug_html_chunked/ChAdvFollowStreamSection.html).
