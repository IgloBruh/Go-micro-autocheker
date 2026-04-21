# AutoCheck Microservices

Готовый вариант исходного проекта, переписанный в микросервисную схему:

`client (HTTP) -> api gateway (gRPC) -> Redis queue -> worker <- MinIO`

и обратно:

`worker -> Redis pub/sub -> api gateway -> client`

## Что внутри

- **gateway** — внешний HTTP API для клиента и внутренний gRPC сервис.
- **worker** — запускает пользовательский код на тестах, сравнивает ответы и публикует результат.
- **Redis** — очередь заданий, канал событий и оперативное состояние запуска.
- **MinIO** — хранение тестов, отправленного кода и итоговых артефактов.

## Поток обработки

1. Клиент делает `POST /api/v1/runs`.
2. HTTP слой gateway вызывает внутренний gRPC метод `SubmitRun`.
3. Gateway сохраняет исходный код в MinIO, ставит задачу в Redis queue и сохраняет статус `PENDING`.
4. Worker берёт задачу из очереди, скачивает код и тесты из MinIO, запускает проверки.
5. Worker сохраняет итог в MinIO и публикует событие завершения в Redis pub/sub.
6. Gateway обновляет состояние запуска.
7. Клиент читает статус через `GET /api/v1/runs/{id}`.

## Быстрый старт

```bash
docker compose up --build
```

После старта:

- HTTP API gateway: `http://localhost:8080`
- gRPC gateway: `localhost:9090`
- MinIO console: `http://localhost:9001`
- MinIO S3 API: `http://localhost:9000`

## Примеры

Создать запуск:

```bash
curl -X POST http://localhost:8080/api/v1/runs \
  -H 'Content-Type: application/json' \
  -d '{
    "task": 0,
    "timeout_ms": 1000,
    "code": "a,b=map(int,input().split());print(a+b)"
  }'
```

Получить результат:

```bash
curl http://localhost:8080/api/v1/runs/<run_id>
```

## Формат ответа

```json
{
  "id": "<run_id>",
  "status": "DONE",
  "task": 0,
  "results": [
    {
      "test_num": 0,
      "status": "OK",
      "time_ms": 12,
      "input_file": "input0.txt",
      "output": "3\n",
      "error": ""
    }
  ]
}
```

## Ограничения выполнения

В worker код запускается как отдельный процесс `python3 -u -c <code>` с таймаутом на каждый тест. Это проще и стабильнее для docker-compose окружения, чем docker-in-docker.

## Локальная проверка без Docker

Нужны:

- Redis
- MinIO / любой S3-compatible storage
- Go 1.22+
- Python 3

```bash
go run ./cmd/gateway
go run ./cmd/worker
```

## Структура

- `cmd/gateway` — запуск gateway
- `cmd/worker` — запуск worker
- `internal/service` — gRPC сервис gateway
- `internal/httpapi` — HTTP API клиента
- `internal/queue` — очередь Redis
- `internal/storage` — MinIO
- `internal/runner` — логика запуска и анализа тестов
- `testdata/minio/tests` — тесты, автоматически загружаемые в MinIO при старте gateway
