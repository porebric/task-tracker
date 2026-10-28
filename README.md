# task-tracker

REST API для управления задачами внутри команд: регистрация и JWT, команды с ролями
owner / admin / member, задачи с историей изменений и комментариями, кеш списка задач в Redis и
SQL-отчёт по команде одним запросом.

Стек: Go 1.26, MySQL 8, Redis 7, Docker Compose, OpenAPI 3 + Swagger UI. Сервис собран на
[healthstep-libs](https://github.com/porebric/healthstep-libs): `app` (fx-каркас, конфиги, логгер,
healthcheck, метрики), `api` (HTTP-слой на chi, OpenAPI, лимитер), `configs`, `logger`.

## Запуск

```bash
make up          # cp .env.example .env (если нет) + docker compose up --build -d --wait
make logs        # логи приложения
make down        # остановить
```

| Что | Адрес |
|---|---|
| API | http://localhost:8080/api/v1 |
| Swagger UI | http://localhost:8080/docs |
| OpenAPI JSON | http://localhost:8080/openapi.json |
| Health (`/live`, `/ready`) | http://localhost:8081 |
| Метрики Prometheus | http://localhost:8082/metrics |

Хост-порты MySQL и Redis задаются в `.env` (`MYSQL_PORT`, `REDIS_PORT`), если стандартные заняты.

Без Docker для приложения (MySQL и Redis уже запущены):

```bash
docker compose up -d --wait mysql redis
make run         # DB_DSN, REDIS_ADDR, JWT_SECRET можно переопределить: make run DB_DSN=...
```

## Конфигурация

Хардкода нет: ключи объявлены в `config/configs_keys.yml`, значения берутся из переменных окружения
(`source: env`, имя переменной — ключ в верхнем регистре) или из `config/configs.yml`
(`source: yaml`). Docker Compose передаёт переменные из `.env`.

| Переменная | Назначение | Default |
|---|---|---|
| `SERVICE_NAME` | имя сервиса (обязательна для `app`) | — |
| `DB_DSN` | DSN MySQL, формат go-sql-driver: `user:pass@tcp(host:3306)/db` | — (обязательна) |
| `DB_MIGRATIONS_PATH` | каталог SQL-миграций, применяются на старте | `./migrations` |
| `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` | подключение к Redis | `127.0.0.1:6379`, пусто, `0` |
| `JWT_SECRET` | секрет подписи HS256 | — (обязательна) |
| `API_PORT`, `HEALTH_PORT`, `METRICS_PORT` | порты HTTP-серверов | `8080`, `8081`, `8082` |
| `API_RATE_LIMIT`, `API_RATE_WINDOW` | лимит запросов с одного IP за окно | `120`, `1m` |
| `LOG_LEVEL`, `LOG_PLAIN` | уровень и формат логов (JSON или текст) | `info`, `false` |
| `TRACER_ENABLED` | OpenTelemetry-трейсинг (`tracer_host`, `tracer_port` по конвенции `app`) | `false` |

В `config/configs.yml`: таймауты HTTP, размеры пула MySQL, порог медленных запросов, TTL кеша
списка задач (`cache_task_list_ttl`, 5m), время жизни токена (`jwt_ttl`, 24h), стоимость bcrypt.

## Миграции

`migrations/*.sql` применяются при старте сервиса в лексикографическом порядке, применённые версии
хранятся в таблице `schema_migrations`. Новая миграция — следующий файл `000N_name.sql`; внутри
обычные SQL-выражения через `;`. Прогнать миграции отдельно от сервиса можно той же командой запуска:
сервис применит их и продолжит работу.

Схема: `users`, `teams`, `team_members` (PK `team_id + user_id`, роль), `tasks` (`status`,
`assignee_id`, `closed_at`, `version`), `task_history` (`changes` JSON), `task_comments`.

## API

Все маршруты, кроме `register` и `login`, требуют заголовок `Authorization: Bearer <token>`.
Полное описание запросов, ответов и ошибок — в Swagger UI.

| Метод и путь | Что делает |
|---|---|
| `POST /api/v1/register` | регистрация |
| `POST /api/v1/login` | выдача JWT |
| `POST /api/v1/teams` | создать команду, текущий пользователь — owner |
| `GET /api/v1/teams` | команды пользователя с его ролью |
| `POST /api/v1/teams/{id}/invite` | добавить участника (`user_id` или `email`, роль `member`/`admin`) |
| `GET /api/v1/teams/{team_id}/stats` | отчёт по команде, owner и admin |
| `POST /api/v1/tasks` | создать задачу в команде |
| `GET /api/v1/tasks?team_id=1&status=todo&assignee_id=5&limit=20&offset=0` | список задач команды |
| `GET /api/v1/tasks/{id}` | задача |
| `PUT /api/v1/tasks/{id}` | обновить задачу (частично) с проверкой прав и записью истории |
| `GET /api/v1/tasks/{id}/history` | история изменений |
| `POST /api/v1/tasks/{id}/comments` | добавить комментарий |
| `GET /api/v1/tasks/{id}/comments` | комментарии |

Статусы задач: `todo`, `in_progress`, `done`. Переход в `done` проставляет `closed_at`, обратный
переход очищает его.

### Примеры

```bash
BASE=http://localhost:8080/api/v1

curl -s $BASE/register -d '{"email":"ann@example.com","password":"secret123","name":"Ann"}'
curl -s $BASE/register -d '{"email":"bob@example.com","password":"secret123","name":"Bob"}'

TOKEN=$(curl -s $BASE/login -d '{"email":"ann@example.com","password":"secret123"}' | jq -r .access_token)
AUTH="Authorization: Bearer $TOKEN"

curl -s $BASE/teams -H "$AUTH" -d '{"name":"core"}'
# {"id":1,"name":"core","created_by":1,"created_at":"..."}

curl -s $BASE/teams/1/invite -H "$AUTH" -d '{"email":"bob@example.com","role":"member"}'

curl -s $BASE/tasks -H "$AUTH" -d '{"team_id":1,"title":"Ship it","description":"...","assignee_id":2}'
# {"id":1,"team_id":1,"title":"Ship it","status":"todo","assignee_id":2,"version":1,...}

curl -s "$BASE/tasks?team_id=1&status=todo&limit=20&offset=0" -H "$AUTH"
# {"tasks":[...],"total":1,"limit":20,"offset":0}

curl -s -X PUT $BASE/tasks/1 -H "$AUTH" -d '{"status":"in_progress","version":1}'
curl -s $BASE/tasks/1/history -H "$AUTH"
# {"history":[{"changes":{"status":{"old":"todo","new":"in_progress"}},"changed_by":1,...}, ...]}

curl -s $BASE/tasks/1/comments -H "$AUTH" -d '{"content":"On it"}'
curl -s $BASE/teams/1/stats -H "$AUTH"
```

Пагинация: `limit` (1..100, по умолчанию 20) и `offset`. Вместо `offset` можно передать
`cursor=<next_cursor>` из предыдущего ответа — курсорная пагинация по `id`, стабильная при
вставках.

### Ошибки

Единый формат: `{"code":104,"message":"...","values":{"field":"why"}}`. Коды 1–8 зарезервированы
HTTP-слоем (1 internal, 2 invalid request, 3 unauthorized, 4 forbidden, 5 not found, 6 conflict,
8 too many requests), коды сервиса — в `pkg/api/codes.go`: 100 email занят, 101 неверные
учётные данные, 102 пользователь не найден, 103 команда не найдена, 104 не участник команды,
105 роль не позволяет, 106 уже участник, 107 задача не найдена, 108 исполнитель не в команде,
109 конфликт версий, 110 поле недоступно для редактирования (список полей — в `values.fields`).

## Роли и права

| Роль | Права |
|---|---|
| owner | всё в рамках команды: приглашать участников с ролью `member` или `admin`, редактировать любые задачи, смотреть отчёт |
| admin | приглашать участников с ролью `member`, редактировать любые задачи, смотреть отчёт |
| member | видеть команду, задачи, историю и комментарии; создавать задачи; редактировать задачи, где он создатель или исполнитель |

Правила редактирования `PUT /tasks/{id}` (меняются только присланные поля):

- owner и admin меняют любые поля любой задачи команды;
- создатель задачи меняет название, описание, статус и исполнителя;
- исполнитель меняет только статус; попытка сменить другие поля — `403`, код 110;
- остальные участники не редактируют чужие задачи;
- исполнитель обязан быть участником команды задачи (`400`, код 108), `assignee_id: 0` снимает
  исполнителя;
- пользователь не видит чужие команды и задачи: чужая команда — `403`, код 104.

Роль owner через приглашение не выдаётся. Права проверяются в сервисном слое
(`internal/service/access.go`), они покрыты unit-тестами.

## Надёжность

**Транзакции.** Создание команды и запись владельца в `team_members`, создание задачи и первая
запись истории, обновление задачи и запись истории — по одной транзакции на операцию.

**Конкурентные обновления.** Обновление берёт строку задачи `SELECT ... FOR UPDATE`, считает diff от
актуального состояния, пишет изменения с `version = version + 1` и запись истории со старыми и
новыми значениями, автором и временем. Параллельные редакторы выстраиваются в очередь, ни одно
изменение не теряется, история всегда строится от последнего зафиксированного состояния.
Дополнительно клиент может прислать `version`: если задача уже изменилась — `409`, код 109.

**Кеш списка задач.** `GET /tasks` кешируется в Redis на `cache_task_list_ttl` (5 минут), ключ
включает команду и все фильтры (`status`, `assignee_id`, `limit`, `offset`, `cursor`). Ключи живут
под «поколением» команды (`tasks:team:{id}:generation`): любое создание или изменение задачи
команды меняет поколение, и все закешированные страницы команды становятся недостижимыми без
`SCAN`/`DEL`, доживая свой TTL. Поколение меняется дважды — под блокировкой строки и после коммита,
чтобы читатель, успевший взять новое поколение до коммита, не оставил в кеше устаревшую страницу.
Недоступный Redis не ломает запросы: кеш пропускается, ошибки пишутся в лог, circuit breaker
(3 подряд ошибки → 10 секунд без обращений к Redis) не даёт ждать таймауты на каждом запросе.

**Лимит запросов.** `API_RATE_LIMIT` запросов с одного IP за `API_RATE_WINDOW`, при превышении
`429`.

## SQL-отчёт

`GET /api/v1/teams/{team_id}/stats` — один запрос с CTE (`internal/repo/stats.go`): задачи
команды выбираются один раз, дальше агрегаты по ним:

- `tasks_by_status` — `GROUP BY status`;
- `top_closers` — топ-3 исполнителей по закрытым задачам за последние 30 дней
  (`JOIN users`, `closed_at >= UTC_TIMESTAMP() - INTERVAL 30 DAY`, `ORDER BY ... LIMIT 3`);
- `avg_close_seconds` — `AVG(TIMESTAMPDIFF(SECOND, created_at, closed_at))`;
- `comments.total` и `comments.by_task` — `LEFT JOIN task_comments` + `GROUP BY` по задаче.

Результат собирается в одну строку через `JSON_OBJECTAGG` / `JSON_ARRAYAGG`, никаких N+1.
Доступ: только owner или admin команды, отчёт строго по одной команде.

## Тесты и линтер

```bash
make test               # unit-тесты (права доступа, JWT, middleware, кеш на miniredis, валидация, миграции)
make test-integration   # поднимает MySQL из compose и гоняет интеграционные тесты
make lint               # golangci-lint
```

Интеграционные тесты (`internal/service/integration_test.go`) ждут MySQL по `TEST_MYSQL_DSN`
(по умолчанию `root:root@tcp(127.0.0.1:3306)/task_tracker_test`, база создаётся автоматически) и
пропускаются, если переменная не задана. Покрыто: SQL-отчёт на реальных данных двух команд,
конкурентные обновления одной задачи, правила доступа к задачам, приглашения и роли.

## Структура

```
cmd/task-tracker        точка входа: сборка fx-приложения из модулей
config/                 декларация ключей и yaml-значения конфигов
migrations/             SQL-миграции
internal/domain         сущности, роли, статусы, доменные ошибки
internal/mysql          подключение к MySQL (GORM), миграции, gorm-логгер
internal/repo           репозиторий: запросы, транзакции, SQL-отчёт
internal/auth           JWT, bcrypt, HTTP-middleware аутентификации
internal/cache          Redis-кеш списка задач с поколениями и circuit breaker
internal/service        бизнес-логика и права доступа
internal/handler        HTTP-хендлеры, маршруты, маппинг ошибок
pkg/api                 контракты запросов и ответов (один файл на запрос), коды ошибок
```

Логи — JSON (zerolog) с `trace_id` на каждый запрос; access-лог библиотеки `api` пишет тела
запросов на уровне `info`, для продакшена с реальными паролями стоит поднять `LOG_LEVEL` до `warn`.
