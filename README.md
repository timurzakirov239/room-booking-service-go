# Room Booking Service (Go)

Тестовый backend-проект сервиса бронирования переговорок на Go.

## Что внутри
- материалы задания: `docs/assignment/`
- архитектурный пакет: `docs/architecture/CHIEF_ARCHITECT_PACKET.md`
- обзор проекта: `PROJECT_OVERVIEW.md`
- SQL-миграции: `db/migrations/`
- bootstrap entrypoint: `cmd/api/main.go`

## Цель
Собрать Go-сервис для бронирования переговорок с PostgreSQL, OpenAPI-first подходом и детерминированной генерацией 30-минутных слотов.

## Текущий статус
В репозитории уже собраны основные Slice-слои:
- bootstrap/config/docker foundation;
- PostgreSQL schema/migration;
- repository layer;
- domain/application logic для materialized slots и booking rules;
- dummy auth + JWT + role middleware;
- router/handler wiring;
- router-level tests/scenario coverage для текущей формы приложения.

На 2026-04-02 в этом runtime уже подтверждено:
- `go mod tidy` проходит;
- `go test ./...` проходит;
- `go build ./...` проходит;
- `docker compose config` проходит.

Неподтверждённой остаётся только живая DB-backed runtime validation (`docker compose up --build` / `curl /_info`), потому что текущий runtime не имеет доступа к Docker daemon socket.

## Явный путь запуска
Предполагаемый локальный run path в нормальном Go/Docker runtime:

### Вариант 1: через Docker Compose
```bash
docker compose up --build
```

Ожидаемое поведение:
- PostgreSQL поднимается из `docker-compose.yml`;
- приложение слушает `:8080`;
- `GET /_info` должен отвечать `200 OK`.

Проверка:
```bash
curl http://localhost:8080/_info
```

### Вариант 2: локально через Go
1. Подготовить PostgreSQL и переменные окружения.
2. Запустить приложение:
```bash
go run ./cmd/api
```

Минимально важные env-переменные:
- `DATABASE_URL`
- `DATABASE_MAX_CONNS`
- `HTTP_PORT` (по умолчанию `8080`)
- `JWT_SECRET`
- `JWT_ISSUER`

### Make targets
```bash
make build
make test
make run
make migrate-up
make migrate-down
```

## Архитектурная логика и rationale
Проект сознательно сделан как один Go backend service без микросервисной сложности.

Ключевые решения:
- **OpenAPI-first**: HTTP shape опирается на `docs/assignment/api.yaml`.
- **PostgreSQL-only**: единственная БД, на ней же держатся ограничения консистентности.
- **Materialized slots**: слоты хранятся в БД как записи со стабильными UUID, а не вычисляются заново на каждый запрос.
- **UTC-only**: время нормализуется в UTC, чтобы не размазывать timezone bugs по handler/repo/domain слоям.
- **Thin handlers**: transport слой отвечает за HTTP/auth mapping; бизнес-правила вынесены в `internal/app` и `internal/domain`.
- **Repo boundary**: SQL-доступ изолирован в `internal/repo/postgres`.

Текущая структура по смыслу:
- `cmd/api` — bootstrap и wiring;
- `internal/config` — env config;
- `internal/api` — transport/auth/router/handlers;
- `internal/app` — use-case orchestration;
- `internal/domain` — правила предметной области;
- `internal/repo` — contracts и pg-backed data access;
- `internal/platform` — infra adapters;
- `db/migrations` — schema evolution.

## Что уже реализовано по текущему shape
Реализованы и покрыты текущими сценарными тестами:
- `GET /_info`
- `POST /dummyLogin`
- `GET /rooms/list`
- `POST /rooms/create`
- `POST /rooms/{roomId}/schedule/create`
- `GET /rooms/{roomId}/slots/list`
- `POST /bookings/create`
- `GET /bookings/list`
- `GET /bookings/my`
- `POST /bookings/{bookingId}/cancel`

## Верификация, реально доступная в текущем runtime
В этом runtime подтверждено следующее:
- `repo_access: yes`
- `go_available: yes`
- `docker_available: yes`
- `compose_available: yes`
- `go test ./...`: green
- `go build ./...`: green
- `docker compose config`: green

Не подтверждено только следующее:
- `docker compose up --build`: blocked by Docker daemon socket permissions (`/var/run/docker.sock`)
- живой `curl http://localhost:8080/_info`: pending после запуска контейнеров или локального PostgreSQL

## Известные блокеры / ограничения
1. **DB-backed integration verification pending**
   - compile/test/build уровень уже подтверждён, но живой прогон приложения с PostgreSQL ещё не подтверждён.

2. **Docker runtime permission blocker**
   - `docker compose up --build` в текущем runtime блокируется правами на Docker daemon socket.
   - Это не compile blocker, а permission blocker окружения.

3. **Push status нужно перепроверить отдельно**
   - ветка сейчас синхронна с `origin/main`, но после новых локальных изменений следующий `git push` ещё не выполнялся.

## Ближайший честный следующий шаг
1. Поднять живой runtime одним из двух путей:
```bash
docker compose up --build
curl http://localhost:8080/_info
```
или, если есть локальный PostgreSQL без Docker:
```bash
go run ./cmd/api
curl http://localhost:8080/_info
```

2. После smoke-check обновить `docs/project/STATUS.md`, закоммитить изменения и сделать `git push`.
