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

На 2026-04-02 подтверждено следующее:
- `go mod tidy` проходит;
- `go test ./...` проходит;
- `go build ./...` проходит;
- `docker compose config` проходит;
- `docker compose up --build -d` проходит;
- `GET /_info` отвечает `200 OK`;
- живой end-to-end сценарий проходит: `dummyLogin -> rooms/create -> schedule/create -> slots/list -> bookings/create -> bookings/my -> bookings/list -> cancel`.

Итого: это уже рабочий backend-сервис, с которым можно взаимодействовать как с HTTP API.

## Как поднять сервис

### Вариант 1: через Docker Compose
Самый простой путь:

```bash
cd /home/timur/.openclaw/workspace/projects/room-booking-service-go
docker compose up --build -d
```

Проверка, что всё поднялось:

```bash
docker compose ps
curl http://127.0.0.1:8080/_info
```

Ожидаемый ответ на `/_info`:
```json
{"status":"ok","service":"room-booking-service-go","databaseOk":true}
```

Остановить сервис:

```bash
docker compose down
```

### Вариант 2: локально через Go
Если хочешь запускать без Docker для приложения, но с отдельным PostgreSQL:

```bash
cd /home/timur/.openclaw/workspace/projects/room-booking-service-go
export DATABASE_URL='postgres://postgres:postgres@127.0.0.1:5432/room_booking?sslmode=disable'
export DATABASE_MAX_CONNS=4
export HTTP_PORT=8080
export JWT_SECRET='dev-secret-change-me'
export JWT_ISSUER='room-booking-service-go'
export AUTO_MIGRATE=true
go run ./cmd/api
```

Потом в другом терминале:

```bash
curl http://127.0.0.1:8080/_info
```

### Make targets
```bash
make build
make test
make run
make migrate-up
```

`migrate-down` сейчас не реализован как полноценный rollback-путь и не должен считаться production-ready функцией.

## Как с ним взаимодействовать
Это backend HTTP API. С ним работают обычными HTTP-запросами.

Базовый адрес локально:

```text
http://127.0.0.1:8080
```

### Базовая логика ролей
- `admin` — создаёт переговорки и расписания, может смотреть общий список броней;
- `user` — смотрит переговорки и слоты, создаёт свои брони и отменяет свои брони.

### Минимальный сценарий использования
1. получить токен `admin`;
2. получить токен `user`;
3. создать переговорку;
4. создать расписание для комнаты;
5. получить список слотов на дату;
6. создать бронь на слот;
7. посмотреть свои брони;
8. при необходимости отменить бронь.

### Примеры запросов

#### 1. Проверка здоровья
```bash
curl http://127.0.0.1:8080/_info
```

#### 2. Получить тестовый токен администратора
```bash
curl -X POST http://127.0.0.1:8080/dummyLogin \
  -H 'Content-Type: application/json' \
  -d '{"role":"admin"}'
```

#### 3. Получить тестовый токен пользователя
```bash
curl -X POST http://127.0.0.1:8080/dummyLogin \
  -H 'Content-Type: application/json' \
  -d '{"role":"user"}'
```

Оба запроса возвращают JSON с полем `token`.

#### 4. Создать комнату (только admin)
```bash
curl -X POST http://127.0.0.1:8080/rooms/create \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ADMIN_TOKEN>' \
  -d '{"name":"Atlas","description":"Переговорка на 6 человек","capacity":6}'
```

#### 5. Получить список комнат
```bash
curl http://127.0.0.1:8080/rooms/list \
  -H 'Authorization: Bearer <USER_OR_ADMIN_TOKEN>'
```

#### 6. Создать расписание для комнаты (только admin)
```bash
curl -X POST http://127.0.0.1:8080/rooms/<ROOM_ID>/schedule/create \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ADMIN_TOKEN>' \
  -d '{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}'
```

`daysOfWeek` — это дни недели от `1` до `7`.

#### 7. Получить слоты на дату
```bash
curl 'http://127.0.0.1:8080/rooms/<ROOM_ID>/slots/list?date=2026-04-03' \
  -H 'Authorization: Bearer <USER_TOKEN>'
```

#### 8. Создать бронь
```bash
curl -X POST http://127.0.0.1:8080/bookings/create \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <USER_TOKEN>' \
  -d '{"slotId":"<SLOT_ID>","createConferenceLink":true}'
```

#### 9. Посмотреть свои брони
```bash
curl http://127.0.0.1:8080/bookings/my \
  -H 'Authorization: Bearer <USER_TOKEN>'
```

#### 10. Посмотреть общий список броней (только admin)
```bash
curl 'http://127.0.0.1:8080/bookings/list?page=1&pageSize=20' \
  -H 'Authorization: Bearer <ADMIN_TOKEN>'
```

#### 11. Отменить бронь
```bash
curl -X POST http://127.0.0.1:8080/bookings/<BOOKING_ID>/cancel \
  -H 'Authorization: Bearer <USER_TOKEN>'
```

## Практическая памятка
Если нужно быстро понять, что сервис жив и база подключена:

```bash
cd /home/timur/.openclaw/workspace/projects/room-booking-service-go
docker compose up --build -d
curl http://127.0.0.1:8080/_info
```

Если `status=ok` и `databaseOk=true`, backend поднят правильно.

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
