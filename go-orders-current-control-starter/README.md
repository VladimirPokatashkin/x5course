# Backend-сервис заказов

Автор: Покаташкин Владимир

## Описание

REST API сервис для управления заказами на Go. Сервис позволяет создавать заказы из нескольких товаров с расчётом итоговой стоимости, получать заказ по уникальному идентификатору и получать список всех заказов, менять статус заказа с валидацией по правилам жизненного цикла и просматривать историю изменения статусов.

## Архитектура и структура

- `cmd/api/main.go` — точка входа в приложение, сборка зависимостей и запуск HTTP-сервера.
- `internal/order` — структуры для представления сущностей (`Order`, `Item`, `StatusChange`, `Status`), интерфейс `Repository` и предметные ошибки.
- `internal/repository/memory` — репозиторий заказов с инкапсулированным внутренним состоянием и управлением через методы.
- `internal/service` — бизнес-логика сервиса (`OrderService`): валидация позиций, расчёт итоговой стоимости заказа, проверка переходов статусов и ведение истории, а также модульные table-driven тесты (`order_service_test.go`).
- `internal/handler` — HTTP-обработчики на базе маршрутизатора `chi`, преобразование ошибок в HTTP-коды и middleware (структурированное логирование запросов с `request_id` и восстановление от паники `recovery`).
- `docs` — сгенерированная документация Swagger.

## Запуск

Требования: Go 1.27+. Все команды выполняются из корня проекта.

Запуск сервера (по умолчанию слушает `:8080`, можно переопределить через переменную окружения `HTTP_ADDR`):

```bash
go mod tidy
go run ./cmd/api
```

## Сборка

```bash
go build ./...
```

После запуска Swagger UI доступен по адресу:
`http://localhost:8080/swagger/index.html`

## Запуск тестов и проверок

```bash
go test ./...
go test ./internal/service/... -cover
go vet ./...
```

Или одной командой через Makefile:

```bash
make check
```

## Доступные эндпоинты

| Метод | Путь | Описание | Код успеха |
|---|---|---|---:|
| `GET` | `/health` | Проверка доступности сервиса | `200 OK` |
| `POST` | `/api/v1/orders` | Создание нового заказа | `201 Created` |
| `GET` | `/api/v1/orders` | Получение списка всех заказов | `200 OK` |
| `GET` | `/api/v1/orders/{id}` | Получение заказа по ID | `200 OK` |
| `PATCH` | `/api/v1/orders/{id}/status` | Изменение статуса заказа | `200 OK` |
| `GET` | `/api/v1/orders/{id}/history` | Получение истории изменения статуса | `200 OK` |
| `GET` | `/swagger/*` | Документация Swagger UI | `200 OK` |

## Примеры запросов

### Создание заказа

```bash
curl -i -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"items":[{"name":"fpga","quantity":1,"price":5000000}]}'
```

### Изменение статуса заказа

Допустимые переходы: 

`new -> processing`

`new -> cancelled`

`processing -> completed`

`processing -> cancelled`

```bash
curl -i -X PATCH http://localhost:8080/api/v1/orders/12/status \
  -H "Content-Type: application/json" \
  -d '{"status":"processing"}'
```

### Получение истории статусов

```bash
curl -i http://localhost:8080/api/v1/orders/12/history
```

## Формат ошибок и HTTP-коды

Все ошибки возвращаются в едином формате JSON:

```json
{
  "error": "invalid status transition",
  "code": "INVALID_STATUS_TRANSITION"
}
```

- `400 Bad Request` — некорректный JSON, невалидный ID, пустой список позиций, некорректные `name`, `quantity`, `price` или неизвестный статус.
- `404 Not Found` — заказ с указанным ID не найден.
- `409 Conflict` — недопустимый переход между статусами заказа.
- `500 Internal Server Error` — внутренняя ошибка сервера.