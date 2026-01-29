# CLAUDE.md

Монорепо для изучения инфраструктурных паттернов: микросервисы, gRPC, Outbox, Kubernetes.

## Структура проекта

```
infr_training/
├── todo_api/          # REST API сервис для управления задачами
├── notify_api/        # gRPC сервис нотификаций (в разработке)
├── pkg/               # Общие пакеты (logger, errors)
└── go.work            # Go workspaces для монорепо
```

## Сервисы

### todo_api (порт 8080)
REST API для CRUD операций с задачами.
- **Стек:** Chi router, PostgreSQL (pgx), Redis (кеширование)
- **Архитектура:** Clean Architecture (ports/adapters)
- **Эндпоинты:** `POST/GET/DELETE /api/v1/task`
- **Подробности:** см. `todo_api/CLAUDE.md`

### notify_api (в разработке)
gRPC сервис для обработки нотификаций.
- **Стек:** gRPC, protobuf
- **Назначение:** получает события от todo_api при создании/удалении задач

## Взаимодействие сервисов

```
[Клиент] --REST--> [todo_api] --gRPC--> [notify_api]
                        │
                        └── PostgreSQL, Redis
```

**Планируемый паттерн:** Outbox для надёжной доставки событий.

## Быстрый старт

```bash
# Поднять инфраструктуру (postgres, redis)
cd todo_api && docker-compose up -d db redis migrate

# Запустить todo_api
export DSN_STRING="postgres://user:password@localhost:5432/todo_store?sslmode=disable"
export REDIS_ADDR="localhost:6379"
go run todo_api/cmd/main.go
```

## Общие команды

```bash
# Сборка всех модулей
go work sync

# Тесты
go test ./...

# Линтер
golangci-lint run ./...
```

## Технологический стек

| Компонент | Технология |
|-----------|------------|
| Язык | Go 1.24 |
| HTTP Router | chi/v5 |
| База данных | PostgreSQL 16 |
| Кеш | Redis 7 |
| Межсервисное взаимодействие | gRPC |
| Контейнеризация | Docker, docker-compose |

## Roadmap

1. ✅ todo_api с PostgreSQL и Redis
2. 🔄 gRPC взаимодействие todo_api → notify_api
3. ⏳ Outbox pattern для надёжной доставки событий
4. ⏳ Kubernetes манифесты
5. ⏳ CI/CD (GitHub Actions)

## Конвенции

- **Архитектура:** Clean Architecture (entities → usecases → ports/adapters)
- **Ошибки:** Wrap через `fmt.Errorf("%w", err)`, общие ошибки в `pkg/errors`
- **Логирование:** Structured logging через zap
- **Конфиг:** Отдельный `internal/config` в каждом сервисе
