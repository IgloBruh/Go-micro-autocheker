# 🚀 AutoCheck Microservices

!\[Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
!\[Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)
!\[Redis](https://img.shields.io/badge/Redis-Queue-red?logo=redis)
!\[MinIO](https://img.shields.io/badge/MinIO-S3--storage-orange)
!\[License](https://img.shields.io/badge/license-MIT-green)

Микросервисная система для автоматической проверки пользовательского
кода на тестах.

\---

## 🧩 Архитектура

``` mermaid
flowchart LR
    Client\[Client HTTP] --> Gateway\[API Gateway]
    Gateway --> RedisQueue\[Redis Queue]
    RedisQueue --> Worker\[Worker]
    Worker --> MinIO\[MinIO Storage]

    Worker --> RedisPubSub\[Redis Pub/Sub]
    RedisPubSub --> Gateway
    Gateway --> Client
```

\---

## ⚙️ Компоненты

### 🔹 Gateway

* HTTP API для клиентов\\
* gRPC сервис для внутреннего взаимодействия\\
* Управление задачами и статусами

### 🔹 Worker

* Выполнение пользовательского кода\\
* Прогон тестов\\
* Сравнение результатов

### 🔹 Redis

* Очередь задач\\
* Pub/Sub события\\
* Хранение состояния

### 🔹 MinIO

* Хранение:

  * тестов\\
  * пользовательского кода\\
  * результатов

\---

## 🔄 Поток обработки

1. `POST /api/v1/runs` --- создание запуска\\
2. Gateway:

   * вызывает `SubmitRun` (gRPC)
   * сохраняет код в MinIO
   * кладёт задачу в Redis
   * статус → `PENDING`
3. Worker:

   * берёт задачу
   * скачивает код и тесты
   * выполняет проверки
4. После выполнения:

   * сохраняет результат в MinIO
   * отправляет событие в Redis Pub/Sub
5. Gateway обновляет статус
6. `GET /api/v1/runs/{id}` --- получение результата

\---

## 🚀 Быстрый старт

``` bash
docker compose up --build
```

\---

