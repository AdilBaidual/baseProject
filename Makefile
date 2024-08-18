# Makefile

# Переменные
PROJECT_NAME := baseProject

PROTO_API := ./api
PROTO_SRC := $(PROTO_API)/$(PROJECT_NAME)/*/*.proto
PROTO_OUT := ./internal/pb


# Вывовд переменных
print-env:
	@echo "PROJECT_NAME = $(PROJECT_NAME)"
	@echo "PROTO_API = $(PROTO_API)"
	@echo "PROTO_SRC = $(PROTO_SRC)"
	@echo "PROTO_OUT = $(PROTO_OUT)"


# Команды для генерации gRPC кода
generate: .generate

.generate:
	@protoc -I $(PROTO_API) \
		--grpc-gateway_out=./internal/pb \
		--grpc-gateway_opt=paths=source_relative \
        --grpc-gateway_opt=generate_unbound_methods=true \
		--go_out=$(PROTO_OUT) \
    	--go_opt=paths=source_relative \
    	--go-grpc_out=$(PROTO_OUT) \
    	--go-grpc_opt=paths=source_relative \
    	--openapiv2_out=./internal/pb \
        --openapiv2_opt=use_go_templates=true \
    	--proto_path=. \
    	$(PROTO_SRC)


# Запуск проекта в docker-compose c тестовыми переменными
run: .test-run

.test-run:
	@docker-compose --env-file test.env up --build