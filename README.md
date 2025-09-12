# Внутренняя очередь задач на Go (in-memory)

Минимальный сервис очереди задач на Go 1.24 без внешних зависимостей. Поддерживает приём задач через HTTP, пул воркеров, ретраи с экспоненциальным бэкоффом и корректное завершение.

## Архитектура
- `internal/http`: HTTP-сервер и хендлеры (`/enqueue`, `/healthz`).
- `internal/config`: загрузка конфигурации из env.
- `internal/queue`: модель `Task`, in-memory `Store`, очередь (канал), воркеры, бэкофф, утилиты.
- `cmd/server`: точка входа, инициализация конфигурации, очереди, воркеров, graceful shutdown.

## Конфигурация (env)
- `WORKERS` — число воркеров (по умолчанию 4, минимум 1).
- `QUEUE_SIZE` — размер буферизированного канала очереди (по умолчанию 64, минимум 1).

## Запуск
```bash
go run ./cmd/server
```
Сервер слушает `:8080`.

## HTTP API
- `GET /healthz` → `200 OK`, пустое тело
- `POST /enqueue` → `202 Accepted` (или `503 Service Unavailable`, если очередь заполнена или приём остановлен).
  - Тело запроса (JSON):
    ```json
    { "payload": {"any": "json"}, "max_retries": 2 }
    ```
  - Пример ответа (`202`):
    ```json
    { "id": "<task-id>", "status": "queued" }
    ```

Примеры curl:
```bash
curl -s -X GET http://localhost:8080/healthz -i
curl -s -X POST http://localhost:8080/enqueue \
  -H 'Content-Type: application/json' \
  -d '{"payload":{"k":"v"},"max_retries":2}' -i
```

## Обработка и ретраи
- Воркеры читают задачи из очереди и обновляют статусы: `queued` → `running` → `done/failed`.
- Ошибки симулируются с вероятностью ~20%.
- При ошибке и наличии попыток выполняется экспоненциальный бэкофф: `delay = base * 2^attempt + jitter`.
  - `base = 200ms`, `jitter ∈ [0..100ms]`.
  - Повторная постановка выполняется неблокирующе, с учётом контекста завершения.

## Допущения
- In-memory хранилище `Store` (нет персистентности), данные теряются при перезапуске.
- Нет аутентификации, троттлинга, backpressure за пределами размера канала.
- Демонстрационная реализация для учебных и тестовых целей.

## Тестирование
Запуск тестов:
```bash
go test ./... -race -coverprofile=coverage.out -covermode=atomic
```

CI: GitHub Actions (`.github/workflows/ci.yml`) на `push`/`pull_request`.
