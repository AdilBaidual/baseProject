# Makefile

# Переменные
PROTO_SRC := api/*/*.proto
PROTO_OUT := ./internal/pb

# Команды для генерации gRPC кода
generate: .generate
.generate:
	@protoc $(PROTO_SRC) \
		--go_out=$(PROTO_OUT) \
    	--go_opt=paths=source_relative \
    	--go-grpc_out=$(PROTO_OUT) \
    	--go-grpc_opt=paths=source_relative \
    	--proto_path=.