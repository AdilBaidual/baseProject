# =============================================================================
# Makefile for baseProject
# =============================================================================

# Project variables
PROJECT_NAME := baseProject
APP_NAME := baseproject
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker variables
DOCKER_IMAGE := $(APP_NAME)

# Protobuf variables
PROTO_API := ./api
PROTO_SRC := $(PROTO_API)/$(PROJECT_NAME)/*/*.proto
PROTO_OUT := ./internal/pb

# Environment files
ENV_LOCAL := env.local

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

# =============================================================================
# Help
# =============================================================================
.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)$(PROJECT_NAME) - Available commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Development:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST) | grep -E "(dev|run|test|generate|clean)"
	@echo ""
	@echo "$(YELLOW)Docker Local:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST) | grep -E "(local|docker-)"
	@echo ""
	@echo "$(YELLOW)Utilities:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST) | grep -E "(lint|fmt|mod|version|info|check|install|env|db-)"

# =============================================================================
# Development
# =============================================================================
.PHONY: run
run: ## Run the application locally
	@echo "$(BLUE)Running application locally...$(RESET)"
	go run ./cmd/main.go

.PHONY: test
test: ## Run tests
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "$(BLUE)Generating coverage report...$(RESET)"
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

.PHONY: generate
generate: ## Generate gRPC code from protobuf files
	@echo "$(BLUE)Generating gRPC code...$(RESET)"
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
	@echo "$(GREEN)gRPC code generation completed$(RESET)"

.PHONY: mod
mod: ## Download and tidy Go modules
	@echo "$(BLUE)Tidying Go modules...$(RESET)"
	go mod download
	go mod tidy
	go mod verify

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(RESET)"
	go fmt ./...
	goimports -w .

.PHONY: lint
lint: ## Run linters
	@echo "$(BLUE)Running linters...$(RESET)"
	golangci-lint run
	@echo "$(GREEN)✓ Go code passed all linter checks$(RESET)"

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	go clean -cache -modcache -testcache
	rm -f coverage.out coverage.html
	rm -rf bin/ dist/

# =============================================================================
# Docker Local
# =============================================================================
.PHONY: local-up
local-up: ## Start local environment with docker-compose
	@echo "$(BLUE)Starting local environment...$(RESET)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Creating .env from $(ENV_LOCAL)...$(RESET)"; \
		cp $(ENV_LOCAL) .env; \
	fi
	docker-compose -f docker-compose.local.yml --env-file .env up -d --build
	@echo "$(GREEN)Local environment started$(RESET)"
	@echo "API: http://localhost:11066"
	@echo "Jaeger: http://localhost:16686"

.PHONY: local-down
local-down: ## Stop local environment
	@echo "$(BLUE)Stopping local environment...$(RESET)"
	docker-compose -f docker-compose.local.yml down

.PHONY: local-logs
local-logs: ## Show logs from local environment
	docker-compose -f docker-compose.local.yml logs -f

.PHONY: local-status
local-status: ## Show status of local services
	@echo "$(BLUE)Local services status:$(RESET)"
	docker-compose -f docker-compose.local.yml ps

.PHONY: local-clean
local-clean: ## Clean local Docker artifacts
	@echo "$(BLUE)Cleaning local Docker artifacts...$(RESET)"
	docker-compose -f docker-compose.local.yml down --volumes --remove-orphans

.PHONY: local-restart
local-restart: local-down local-up ## Restart local environment

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image: $(DOCKER_IMAGE):latest$(RESET)"
	docker build -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)Docker image built successfully$(RESET)"

# =============================================================================
# Development/Production Compose (for testing)
# =============================================================================
.PHONY: dev-up
dev-up: ## Start development environment (requires .env)
	@echo "$(BLUE)Starting development environment...$(RESET)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Error: .env file not found. Create it based on examples in GITLAB-VARIABLES.md$(RESET)"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.dev.yml --env-file .env up -d --build

.PHONY: dev-down
dev-down: ## Stop development environment
	docker-compose -f docker-compose.dev.yml down

.PHONY: prod-up
prod-up: ## Start production environment (requires .env)
	@echo "$(BLUE)Starting production environment...$(RESET)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Error: .env file not found. Create it based on examples in GITLAB-VARIABLES.md$(RESET)"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.prod.yml --env-file .env up -d --build

.PHONY: prod-down
prod-down: ## Stop production environment
	docker-compose -f docker-compose.prod.yml down

# =============================================================================
# Testing & Health Checks
# =============================================================================
.PHONY: api-test
api-test: ## Test API endpoints
	@echo "$(BLUE)Testing API endpoints...$(RESET)"
	@curl -X POST http://localhost:11066/test && echo ""

.PHONY: health-check
health-check: ## Check if local API is responding
	@echo "$(BLUE)Checking API health...$(RESET)"
	@if curl -f -X POST http://localhost:11066/test > /dev/null 2>&1; then \
		echo "$(GREEN)✓ API is responding$(RESET)"; \
	else \
		echo "$(RED)✗ API is not responding$(RESET)"; \
	fi

# =============================================================================
# Database
# =============================================================================
.PHONY: db-migrate
db-migrate: ## Run database migrations (local)
	@echo "$(BLUE)Running database migrations...$(RESET)"
	@if command -v goose >/dev/null 2>&1; then \
		cd migrations && goose postgres "$(shell grep DATABASE_URL env.local | cut -d'=' -f2)" up; \
		echo "$(GREEN)✓ Database migrations completed$(RESET)"; \
	else \
		echo "$(RED)goose not found. Install with: go install github.com/pressly/goose/v3/cmd/goose@latest$(RESET)"; \
	fi

.PHONY: db-create-migration
db-create-migration: ## Create new migration file (usage: make db-create-migration NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)❌ Migration name is required$(RESET)"; \
		echo "$(YELLOW)Usage: make db-create-migration NAME=your_migration_name$(RESET)"; \
		echo "$(CYAN)Example: make db-create-migration NAME=add_users_table$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating new migration: $(NAME)$(RESET)"
	@if command -v goose >/dev/null 2>&1; then \
		cd migrations && goose create $(NAME) sql; \
		echo "$(GREEN)✓ Migration created successfully$(RESET)"; \
		echo "$(CYAN)Files created in migrations/ directory$(RESET)"; \
		ls -la migrations/*$(NAME)* 2>/dev/null || echo "$(YELLOW)Check migrations/ directory for new files$(RESET)"; \
	else \
		echo "$(RED)goose not found. Install with: go install github.com/pressly/goose/v3/cmd/goose@latest$(RESET)"; \
		exit 1; \
	fi

.PHONY: db-status
db-status: ## Show migration status
	@echo "$(BLUE)Checking migration status...$(RESET)"
	@if command -v goose >/dev/null 2>&1; then \
		cd migrations && goose postgres "$(shell grep DATABASE_URL env.local | cut -d'=' -f2)" status; \
	else \
		echo "$(RED)goose not found. Install with: go install github.com/pressly/goose/v3/cmd/goose@latest$(RESET)"; \
	fi

.PHONY: db-rollback
db-rollback: ## Rollback last migration
	@echo "$(YELLOW)⚠️  Rolling back last migration...$(RESET)"
	@if command -v goose >/dev/null 2>&1; then \
		cd migrations && goose postgres "$(shell grep DATABASE_URL env.local | cut -d'=' -f2)" down; \
		echo "$(GREEN)✓ Migration rollback completed$(RESET)"; \
	else \
		echo "$(RED)goose not found. Install with: go install github.com/pressly/goose/v3/cmd/goose@latest$(RESET)"; \
	fi

.PHONY: db-connect
db-connect: ## Connect to database (local)
	@echo "$(BLUE)Connecting to local database...$(RESET)"
	docker-compose -f docker-compose.local.yml exec postgres psql -U admin -d admin

.PHONY: db-logs
db-logs: ## Show database logs
	docker-compose -f docker-compose.local.yml logs postgres

# =============================================================================
# Utilities
# =============================================================================
.PHONY: version
version: ## Show version information
	@echo "$(CYAN)Version Information:$(RESET)"
	@echo "  Project:     $(PROJECT_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Git Commit:  $(GIT_COMMIT)"

.PHONY: info
info: ## Show environment information
	@echo "$(CYAN)Environment Information:$(RESET)"
	@echo "  Go Version:    $(shell go version)"
	@echo "  Docker:        $(shell docker --version 2>/dev/null || echo 'Not installed')"
	@echo "  Docker Compose: $(shell docker-compose --version 2>/dev/null || echo 'Not installed')"
	@echo "  Project Root:  $(PWD)"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	@echo "$(YELLOW)Installing golangci-lint...$(RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(YELLOW)Installing goimports...$(RESET)"
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(YELLOW)Installing protoc-gen-go...$(RESET)"
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "$(YELLOW)Installing protoc-gen-go-grpc...$(RESET)"
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "$(YELLOW)Installing protoc-gen-grpc-gateway...$(RESET)"
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "$(YELLOW)Installing protoc-gen-openapiv2...$(RESET)"
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "$(YELLOW)Installing goose for database migrations...$(RESET)"
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "$(GREEN)✓ All development tools installed$(RESET)"
	@echo "$(CYAN)Note: Make sure protoc is installed:$(RESET)"
	@echo "  macOS: brew install protobuf"
	@echo "  Ubuntu: sudo apt install protobuf-compiler"
	@echo "  Arch: sudo pacman -S protobuf"

.PHONY: check-tools
check-tools: ## Check if all required tools are installed
	@echo "$(BLUE)Checking required tools...$(RESET)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)✗ Go not found$(RESET)" && exit 1)
	@command -v golangci-lint >/dev/null 2>&1 || echo "$(YELLOW)⚠ golangci-lint not found (run: make install-tools)$(RESET)"
	@command -v goimports >/dev/null 2>&1 || echo "$(YELLOW)⚠ goimports not found (run: make install-tools)$(RESET)"
	@command -v protoc >/dev/null 2>&1 || echo "$(YELLOW)⚠ protoc not found (install: brew install protobuf)$(RESET)"
	@command -v protoc-gen-go >/dev/null 2>&1 || echo "$(YELLOW)⚠ protoc-gen-go not found (run: make install-tools)$(RESET)"
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || echo "$(YELLOW)⚠ protoc-gen-go-grpc not found (run: make install-tools)$(RESET)"
	@command -v goose >/dev/null 2>&1 || echo "$(YELLOW)⚠ goose not found (run: make install-tools)$(RESET)"
	@command -v docker >/dev/null 2>&1 || echo "$(YELLOW)⚠ Docker not found$(RESET)"
	@command -v docker-compose >/dev/null 2>&1 || echo "$(YELLOW)⚠ docker-compose not found$(RESET)"
	@echo "$(GREEN)✓ All required tools are installed$(RESET)"

.PHONY: env-setup
env-setup: ## Setup .env file for local development
	@if [ ! -f .env ]; then \
		echo "$(BLUE)Creating .env file from $(ENV_LOCAL)...$(RESET)"; \
		cp $(ENV_LOCAL) .env; \
		echo "$(GREEN).env file created$(RESET)"; \
		echo "$(YELLOW)You can modify .env file for your local setup$(RESET)"; \
	else \
		echo "$(YELLOW).env file already exists$(RESET)"; \
	fi

# =============================================================================
# Default target
# =============================================================================
.DEFAULT_GOAL := help
