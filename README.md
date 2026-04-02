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
- начальная PostgreSQL schema/migration;
- repository layer;
- domain/application logic для materialized slots и booking rules;
- dummy auth + JWT + role middleware;
- router/handler wiring;
- router-level tests/scenario coverage для текущей формы приложения.

Это состояние близко к implementation-ready, но финальная runnable validation в текущем runtime ограничена отсутствием `go` и `docker`.

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
Доступно и/или частично реализовано:
- `GET /_info`
- `POST /dummyLogin`
- `GET /rooms/list`
- `POST /rooms/create`
- `POST /bookings/create`
- `GET /bookings/my`
- `POST /bookings/{bookingId}/cancel`

Часть route shape всё ещё scaffold-level:
- `POST /rooms/{roomId}/schedule/create`
- `GET /rooms/{roomId}/slots/list`
- `GET /bookings/list`

## Верификация, реально доступная в текущем runtime
В этом runtime было подтверждено только следующее:
- `repo_access: yes`
- `go_available: no`
- `docker_available: no`
- `compose_available: no`

Следствие:
- `go test ./...` здесь не запускался, потому что `go` недоступен;
- `go build ./...` здесь не запускался, потому что `go` недоступен;
- `docker compose up --build` здесь не запускался, потому что `docker`/`compose` недоступны.

## Известные блокеры / ограничения
1. **Runtime/toolchain blocker**
   - В текущем окружении отсутствуют `go`, `docker` и `docker compose`.
   - Поэтому нельзя честно подтвердить compile/run/integration readiness прямо здесь.

2. **Не полностью завершённые endpoint slices**
   - schedule create / slots list / admin bookings list пока не доведены до полной business-complete реализации.

3. **DB-backed integration verification pending**
   - Router-level tests и сценарные проверки добавлены в кодовую базу, но фактический прогон против реального Go/PostgreSQL runtime остаётся следующим шагом.

4. **Push blocker**
   - Push в origin в этом runtime остаётся заблокирован отсутствием GitHub credentials.

## Ближайший честный следующий шаг
В окружении с установленными Go и Docker выполнить:
```bash
go test ./...
go build ./...
docker compose up --build
curl http://localhost:8080/_info
```

После этого можно закрывать compile/runtime verification и добивать оставшиеся scaffold-level endpoints.
