# Архитектурный пакет — Главный архитектор

Дата: 2026-04-02
Статус: architecture-pass complete
Источник: внутренний handoff от Главного архитектора

## 1. Summary
Рекомендация: реализовать сервис как один Go backend-монолит с чёткими доменными модулями, PostgreSQL как единственную БД, OpenAPI-first разработку от `api.yaml` и детерминированную стратегию генерации слотов на стороне доменного сервиса.

Это должен быть implementation-ready сервис под тестовое задание, а не «платформа».
Главный акцент:
- чистый контракт по `api.yaml`;
- предсказуемые бизнес-правила;
- идемпотентность и валидация;
- понятная схема хранения и генерации слотов;
- хороший локальный запуск через Docker Compose.

## 2. Problem framing
Это не просто CRUD по API, а небольшой transactional backend с календарно-слотовой логикой.
Ключевые архитектурные риски:
- корректная генерация слотов;
- недопущение конфликтов и дублей;
- консистентность статусов;
- согласованность между API-контрактом и доменной моделью;
- предсказуемое поведение при конкурентных запросах.

## 3. Chosen primary architecture team and why
Основная команда: Архитектор систем.

Почему:
- задача в первую очередь про shape backend-service;
- главный вопрос — runtime/service boundaries, API flow, generation flow, consistency, delivery shape;
- нужен buildable сервисный дизайн, а не абстрактная схема.

Поддерживающая команда: Архитектор данных.

Почему:
- слотогенерация, ограничения, статусы и дедупликация упираются в таблицы, индексы, уникальные ограничения, транзакции, миграции и схему PostgreSQL.

## 4. Constraints from assignment/spec
Обязательные ограничения:
- язык: Go;
- база данных: PostgreSQL;
- есть `api.yaml`, значит нужен contract-first/spec-driven подход;
- проект должен запускаться через Docker Compose;
- сервис должен слушать порт 8080;
- нужен `GET /_info`, который всегда возвращает `200 OK`;
- это backend-service test assignment, а не большой продукт.

Практические следствия:
- без message broker;
- без микросервисов;
- без лишней инфраструктурной сложности;
- оптимизация под простую сборку, понятную структуру и воспроизводимый локальный запуск.

## 5. Recommended solution for Go implementation
Рекомендуемая архитектурная форма:
- один Go сервис;
- слоистая, но не переусложнённая структура;
- split: domain / application / infrastructure / api;
- OpenAPI как источник HTTP-контракта;
- PostgreSQL + migrations;
- Docker Compose для local dev;
- генерация слотов как отдельный domain/application use case.

Практический стек:
- HTTP router: `chi`;
- OpenAPI: `oapi-codegen`;
- DB access: `pgx` + `sqlc`;
- migrations: `goose` или `golang-migrate`;
- config: env-based;
- logging: `slog`;
- testing: стандартный `testing` + `testify` + integration tests против PostgreSQL.

## 6. Module / service / data design
Рекомендуемая структура модулей:
- `api` — generated OpenAPI contracts и transport layer;
- `app` — use cases / orchestration;
- `domain` — entities, business rules, slot generation logic;
- `repo` — interfaces и SQL-backed implementations;
- `db` — migrations и queries;
- `config`;
- `cmd/api` — bootstrap.

Ожидаемые доменные сущности:
- user;
- room;
- schedule;
- slot;
- booking.

Принципы data design:
- слот хранит `start/end`, `room_id`, состояние занятости через активную бронь;
- уникальность и консистентность защищаются на уровне PostgreSQL;
- все transition rules оформляются явно в domain/application layer;
- timestamps и вся работа со временем — только в UTC.

## 7. Slot-generation strategy and rationale
Рекомендация: использовать детерминированную rule-based генерацию слотов на основе:
- расписания комнаты;
- дня недели;
- fixed slot duration = 30 minutes;
- fixed generation horizon;
- дедупликации на уровне БД.

Чистая стратегия для этого проекта:
1. расписание создаётся один раз;
2. после создания расписания система материализует слоты на rolling window;
3. при запросе списка слотов система читает уже сохранённые слоты по `roomId + date`, а не генерирует их в хендлере;
4. фоновый refill/ensure-window может быть оформлен как application service либо как синхронный ensure-step при schedule creation / startup.

Почему это лучше, чем «генерировать на лету»:
- стабильные UUID слотов;
- проще бронирование по `slotId`;
- проще выдерживать уникальность;
- легче писать тесты и объяснять систему в README;
- лучше соответствует высоконагруженному endpoint `/rooms/{roomId}/slots/list`.

## 8. API / auth / business-rule risks
API-risks:
- рассинхрон между `api.yaml` и реальным поведением;
- слишком толстые handlers;
- отсутствие чёткой валидации.

Auth-risks:
- раздувание auth сверх задания;
- путаница между dummy auth и optional real auth.

Business-rule risks:
- повторная генерация слотов создаёт дубли;
- конкурентные booking requests ломают консистентность;
- booked slot случайно перегенерируется или становится доступным;
- timezone boundary bugs;
- business rules размазываются между HTTP, SQL и service layer.

Mitigation:
- доменные инварианты в application/domain;
- уникальные ограничения и транзакции в PostgreSQL;
- integration tests на генерацию и конкурентные кейсы;
- строгая UTC-нормализация.

## 9. Testing strategy and delivery shape
Нужно 3 уровня тестов:

### Unit tests
- slot generation logic;
- validation;
- state transitions;
- time-window edge cases.

### Integration tests
- against real PostgreSQL;
- migrations;
- repository behavior;
- uniqueness/conflict behavior;
- generation transactional behavior.

### API / E2E tests
- сценарий: создание переговорки -> создание расписания -> создание брони пользователем;
- сценарий: отмена брони пользователем;
- проверка соответствия `api.yaml` по статусам и payloads;
- негативные кейсы.

Recommended delivery shape:
- один repo;
- `docker-compose.yml` для app + postgres;
- `Makefile`;
- migrations folder;
- `README.md` с описанием запуска, тестов, решений и принятых компромиссов.

## 10. Open questions to Timur only if truly blocking
На архитектурном проходе критически blocking ambiguity нет.

Неблокирующие уточнения, отмеченные архитектором:
1. считать ли `api.yaml` единственным источником API-правды, если текст задания местами мягче;
2. нужен ли только минимальный auth из spec или стоит делать optional `/register` и `/login`;
3. как трактовать повторную генерацию слотов на тот же диапазон, если implementation пойдёт через materialized slots.

## Operational note
Для текущего implementation pass можно стартовать без ожидания ответов на эти 3 вопроса, если в README явно зафиксировать выбранные решения.