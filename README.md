# Engineering Calc API

REST API для инженеров ОВиК: хранит проекты, инженерные системы и участки воздуховодов/трубопроводов
и считает для каждого участка площадь сечения и скорость потока.

## Возможности

- **Проекты** → **системы** (среда: воздух или вода) → **участки** (прямоугольные или круглые).
- Для каждого участка API автоматически считает:
  - площадь сечения, м²;
  - скорость потока, м/с, по формуле `v = L / (3600 · F)`, где `L` — расход в м³/ч, `F` — площадь в м².
- Валидация входных данных на уровне домена (форма, геометрия, длина, расход).
- Кэширование списка участков системы в Redis (TTL 30 с, сбрасывается при создании участка).
- Хранение в PostgreSQL, миграции через [golang-migrate](https://github.com/golang-migrate/migrate).

## Стек

Go · gorilla/mux · PostgreSQL (pgx) · Redis (go-redis) · Docker Compose · testcontainers · miniredis

## Структура

```
cmd/api/                     точка входа, сборка зависимостей
internal/domain/             сущности, бизнес-правила и расчёты (Area, Velocity)
internal/service/            сценарии: проверка родителя, кэш
internal/repository/         in-memory хранилища (используются в тестах)
internal/repository/postgres PostgreSQL-репозитории
internal/repository/redis_store  клиент Redis
internal/transport/http_transport  HTTP-хендлеры, DTO, роутинг
migrations/                  SQL-миграции
```

## Быстрый старт

Нужен только Docker:

```bash
docker compose up -d --build
```

Поднимутся Postgres, Redis, контейнер с миграциями и само API на `http://localhost:9091`.

### Локальный запуск без Docker для API

```bash
docker compose up -d postgres redis migrate
cp .env.example .env
go run ./cmd/api
```

### Переменные окружения

| Переменная     | Обязательна | По умолчанию | Описание                    |
|----------------|-------------|--------------|-----------------------------|
| `DATABASE_URL` | да          | —            | DSN PostgreSQL              |
| `REDIS_ADDR`   | да          | —            | адрес Redis, `host:port`    |
| `HTTP_ADDR`    | нет         | `:9091`      | адрес HTTP-сервера          |

## API

| Метод | Путь                               | Описание                     |
|-------|------------------------------------|------------------------------|
| POST  | `/projects`                        | создать проект               |
| GET   | `/projects`                        | список проектов              |
| GET   | `/projects/{id}`                   | проект по id                 |
| POST  | `/projects/{project_id}/systems`   | создать систему в проекте    |
| GET   | `/projects/{project_id}/systems`   | системы проекта              |
| POST  | `/systems/{system_id}/segments`    | создать участок в системе    |
| GET   | `/systems/{system_id}/segments`    | участки системы (с расчётом) |

Ошибки возвращаются в виде `{"message": "..."}` с кодами `400` (невалидные данные),
`404` (родитель не найден), `409` (дубликат), `500`.

### Пример

```bash
# 1. Проект
curl -X POST localhost:9091/projects \
  -d '{"name":"Офис","description":"Вентиляция офиса"}'

# 2. Система (medium: air | water)
curl -X POST localhost:9091/projects/<project_id>/systems \
  -d '{"name":"П1","medium":"air","purpose":"приток"}'

# 3. Участок: размеры в мм, длина в м, расход в м³/ч
curl -X POST localhost:9091/systems/<system_id>/segments \
  -d '{"name":"Участок 1","shape":"rect","rect":{"width":400,"height":200},"length":5,"flow_rate":1000}'
```

Ответ:

```json
{
  "id": "b3d2578c-...",
  "system_id": "d9312010-...",
  "name": "Участок 1",
  "shape": "rect",
  "rect": { "width": 400, "height": 200 },
  "round": null,
  "length": 5,
  "flow_rate": 1000,
  "area": 0.08,
  "velocity": 3.472,
  "created_at": "10.10.2026"
}
```

Для круглого участка вместо `rect` передаётся `"shape":"round","round":{"diameter":315}`.

### Единицы измерения

| Поле                          | Единица |
|-------------------------------|---------|
| `width`, `height`, `diameter` | мм      |
| `length`                      | м       |
| `flow_rate`                   | м³/ч    |
| `area`                        | м²      |
| `velocity`                    | м/с     |

## Тесты

```bash
# юнит-тесты (domain, service, http, in-memory хранилища)
go test ./...

# интеграционные тесты PostgreSQL-репозиториев (нужен запущенный Docker)
go test -tags integration ./internal/repository/postgres/
```
