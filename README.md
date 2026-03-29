# Go Base Project

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![gRPC](https://img.shields.io/badge/gRPC-Protocol_Buffers-244c5a?style=flat-square&logo=grpc)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?style=flat-square&logo=postgresql)](https://www.postgresql.org/)

Шаблон Go-приложения с gRPC/gRPC-Gateway API, PostgreSQL, Jaeger tracing. Удобная основа для микросервисов.

## Возможности

- gRPC + REST API через gRPC-Gateway
- Protocol Buffers для типобезопасности
- PostgreSQL
- Distributed Tracing через Jaeger
- Dependency Injection через Uber FX
- Swagger UI
- Clean Architecture

## Структура

```
baseProject/
├── api/                     # Proto файлы
│   ├── google/api/          # Google API annotations
│   └── baseProject/test/    # Proto сервиса
│
├── cmd/
│   └── main.go              # Точка входа
│
├── config/
│   ├── config.go            # Загрузка конфигурации
│   └── config.yaml
│
├── internal/
│   ├── app/
│   │   ├── app.go           # Инициализация
│   │   └── test/            # Обработчики
│   │
│   ├── container/
│   │   └── service_container.go
│   │
│   ├── errors/
│   │   └── errors.go
│   │
│   ├── interceptor/
│   │   └── interceptor.go   # gRPC интерцепторы
│   │
│   ├── pb/                  # Сгенерированные protobuf
│   │
│   ├── service/
│   │   └── testservice/     # Бизнес-логика
│   │
│   └── store/               # Репозитории
│
├── pkg/
│   ├── grpcserver/
│   │   └── server.go
│   ├── httpserver/
│   │   └── server.go
│   ├── jaeger/
│   │   └── jaeger.go
│   └── storage/postgres/
│       └── postgres.go
│
├── db/migrations/           # Миграции
├── .gitlab-ci.yml           # CI/CD
├── Makefile                 # Команды
└── docker-compose.*.yml     # Docker окружения
```

## Технические решения

**gRPC + REST Gateway:** один proto-файл генерирует два интерфейса.

**Clean Architecture:** delivery → usecase → repository.

**DI:** Uber FX для управления зависимостями и graceful shutdown.

**Tracing:** Jaeger для отслеживания запросов.

## Запуск

```bash
# Установить зависимости
go mod download

# Установить protoc плагины
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Сгенерировать protobuf
cd api
protoc -I . --go_out ../internal/pb --go_opt paths=source_relative \
  --go-grpc_out ../internal/pb --go-grpc_opt paths=source_relative \
  --grpc-gateway_out ../internal/pb --grpc-gateway_opt paths=source_relative \
  baseProject/test/test.proto
cd ..

# Настроить окружение
cp env.local .env

# Запустить инфраструктуру
docker-compose -f docker-compose.local.yml up -d postgres jaeger

# Запустить приложение
go run cmd/main.go
```

## API

**gRPC:** `localhost:50051`

**REST:**
```
GET http://localhost:8080/v1/ping
GET http://localhost:8080/swagger-ui/
```

## Конфигурация

```yaml
grpc:
  port: 50051

http:
  port: 8080

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: baseproject

jaeger:
  endpoint: http://localhost:14268/api/traces
```

## Makefile

```bash
make build          # Сборка
make test           # Тесты
make migrate-up     # Миграции
make generate       # Генерация protobuf
make lint           # Линтинг
make docker-build   # Docker образ
```

## Как использовать как шаблон

1. Заменить `baseProject` на имя проекта
2. Обновить proto файлы в `api/`
3. Сгенерировать код: `make generate`
4. Реализовать бизнес-логику в `internal/service/`

## Зависимости

- gRPC
- grpc-gateway
- pgx (PostgreSQL)
- Uber FX
- Zap (логирование)
- Jaeger (трейсинг)
