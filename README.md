# itmo-high-load-system-1lab

[Отчет](report.pdf) · [Запуск](#запуск) · [Команды](#команды) · [Устройство проекта](#устройство-проекта)

Go · Prometheus · Grafana · Pandora

[![Tests](https://img.shields.io/github/actions/workflow/status/apartapatia/itmo-y2-s2-high-load-systems/ci.yml?branch=main&event=push&style=flat-square&label=tests&labelColor=101010&color=FFC799&logo=githubactions&logoColor=FFC799)](https://github.com/apartapatia/itmo-y2-s2-high-load-systems/actions/workflows/ci.yml)

---

## Запуск

Все команды выполняются из каталога с `Makefile` и `docker-compose.yml`.
Для запуска нужны Make, Docker Engine и Docker Compose v2. Для проверки HTTP
используется curl. Первая сборка требует доступа к реестрам образов, GitHub и Go реджестри.
Go 1.26.0 на хосте нужен только для локального запуска и тестов.

Создайте файл окружения. Существующий `.env` команда не перезаписывает.

```sh
make docker-init
```

Заполните `GRAFANA_ADMIN_USER` и `GRAFANA_ADMIN_PASSWORD` в `.env`, затем запустите деплой лабораторной.

```sh
make docker-up
make docker-ps
make docker-health
make dashboard-logs
```

Docker compose запускает два экземпляра приложения, Prometheus и Grafana. Генератор
ждёт успешных healthchecks, создаёт дашборд и завершается.
После этого запускается Pandora.
Через несколько сборов метрик, обычно за 2 - 3 минуты после старта нагрузки,
на дашборде появляются данные.

| Компонент | Адрес |
| --- | --- |
| Приложение, инстанс 1 | http://localhost:8080 |
| Приложение, инстанс 2 | http://localhost:8081 |
| Prometheus | http://localhost:9090 |
| Таргеты Prometheus | http://localhost:9090/targets |
| Grafana | http://localhost:3000 |
| Дашборд | http://localhost:3000/d/itmo-high-load-system-1lab |

## Настройки

| Переменная | Где задаётся | Назначение |
| --- | --- | --- |
| `GRAFANA_ADMIN_USER` | `.env`, пример `admin` | Пользователь Grafana и генератора |
| `GRAFANA_ADMIN_PASSWORD` | `.env`, без значения в примере | Пароль Grafana и генератора |
| `POD_ID` | docker compose, `1` и `2` | Идентификатор экземпляра в метриках |
| `DATACENTER` | docker compose, `dc1` и `dc2` | Логический датацентр |
| `APP_URL` | docker compose для генератора | Откуда читать список ручек |
| `GRAFANA_URL` | docker compose для генератора | Куда сохранять дашборд |

## Команды

| Команда | Действие |
| --- | --- |
| `make docker-config` | Проверить docker compose без запуска |
| `make docker-up` | Собрать и запустить лабораторную в фоне |
| `make docker-ps` | Показать работающие и завершившиеся контейнеры |
| `make docker-health` | Проверить оба экземпляра приложения |
| `make docker-logs` | Следить за логами всей лабы |
| `make docker-down` | Остановить лабу, сохранив тома мониторинга |
| `make dashboard-generate` | Повторно сгенерировать дашборд |
| `make dashboard-logs` | Показать логи генератора |
| `make load-logs` | Следить за логами Pandora |
| `make load-stop` | Остановить нагрузку |
| `make load-start` | Запустить нагрузку, остановленный контейнер начнёт разгон заново |
| `make load-smoke` | Остановить фононовую нагрущку и выполнить разгон 5 с с удержанием 15 с |
| `make app-run` | Запустить go сервис локально |
| `make app-test` | Выполнить тесты go |
| `make docker-build` | Собрать образ приложения |
| `make docker-run` | Запустить отдельный контейнер приложения |

## Разработка

Для локального запуска задайте метки экземпляра

```sh
POD_ID=local DATACENTER=local make app-run
```

Порт 8080 должен быть свободен. Ручки регистрируются в общем списке `routes`
в `internal/service/handler.go`, который используется для метрик и `/handlers`.
После добавления маршрута выполните `make docker-up` и
`make dashboard-generate`, затем дополните сценарии Pandora.

## Устройство проекта

```text
cmd/
  server/                 запуск HTTP-сервера и завершение работы
  dashboard-generator/    генерация дашборда и клиент API Grafana
internal/service/         маршруты, хранилище и метрики
load/                     сценарии Pandora, профиль нагрузки и runner
monitoring/               конфигурация Prometheus и datasource Grafana
```

### HTTP-сервис

| Маршрут | Назначение |
| --- | --- |
| `POST /items` | Создать запись с полями `id` и `name`, вернуть 201 или 409 при повторном ID |
| `GET /items/{id}` | Получить запись, вернуть 404 при отсутствии |
| `PUT /items/{id}` | Обновить `name`, сохранив ID из пути, вернуть 404 при отсутствии |
| `GET /health` | Проверка готовности приложения |
| `GET /handlers` | Список маршрутов для генератора |
| `GET /metrics` | Метрики Prometheus |

### Мониторинг

| Метрика | Тип и назначение |
| --- | --- |
| `http_requests_total` | `Counter` для завершённых HTTP запросов и ответов по кодам |
| `http_request_duration_seconds` | `Histogram` для распределения серверных длительностей |

Histogram позволяет объединять распределения разных экземпляров по одинаковым
границам бакетов и вычислять перцентили для выбранных подов.

### Генератор дашборда

Генератор получает список ручек через `/handlers`, формирует описание панелей
и сохраняет дашборд через HTTP API Grafana. Для каждой ручки создаются latency
p50/p95/p99, RPS и ошибки по кодам. Селекторы датацентра и пода поддерживают
несколько значений и All. Редактирование дашборда отключено.

Сверху расположена сводка p50/p95/p99, RPS, частоты и доли ошибок по выбранным
экземплярам. Ниже идут сворачиваемые секции ручек.

UID равен `itmo-high-load-system-1lab`. Повторная генерация обновляет тот же
дашборд целиком. После добавления маршрута пересоберите приложение командой
`make docker-up`, затем при необходимости выполните `make dashboard-generate`.

### Нагрузка

Pandora v0.6.2 работает в двух пулах, по одному на экземпляр приложения.
Параметры `http/scenario` задают частоту сценариев, а не отдельных HTTP запросов.
Профиль каждого пула разгоняется с 1 до 5 сценариев/с за две минуты.
После разгона часовая фоновая фаза повторяется до остановки контейнера.
На границе фаз возможна короткая пауза перезапуска Pandora.

Сценарии охватывают POST, GET и PUT, включая ожидаемые ответы 404 и 409.