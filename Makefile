# Переменные
APP_NAME = TODO_app
DOCKER_IMAGE = TODO_app
GO_FILES = $(shell find . -type f -name '*.go')
PORT = 8080
DB_PORT = 5432
BIN_DIR = bin

.PHONY: all build run test clean docker-build docker-run compose-up compose-down tidy

all: build

# Сборка Go приложения
build: $(BIN_DIR)/$(APP_NAME)

$(BIN_DIR)/$(APP_NAME): $(GO_FILES)
	@echo "Building Go application..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/app/

# Запуск собранного приложения
run: build
	@echo "Starting application..."
	@./$(BIN_DIR)/$(APP_NAME)

# Запуск напрямую через go run (без предварительной сборки)
run-dev:
	@echo "Starting in development mode..."
	@go run ./cmd/app/

# Тесты
test:
	@echo "Running tests..."
	@go test -v ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Docker
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -d -p $(PORT):$(PORT) --name $(APP_NAME) $(DOCKER_IMAGE)

docker-stop:
	@docker stop $(APP_NAME) || true

docker-rm:
	@docker rm $(APP_NAME) || true

# Docker Compose
compose up:
	@echo "Starting services with Docker Compose..."
	@docker compose up -d --build

compose down:
	@echo "Stopping services..."
	@docker compose down

compose logs:
	@docker compose logs -f

# Go модули
tidy:
	@go mod tidy -v

vendOR:
	@go mod vendor

# Очистка
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out
	@go clean -cache -testcache

# Помощь
help:
	@echo "Available targets:"
	@echo "  build        - Build Go application"
	@echo "  run          - Build and run application"
	@echo "  run-dev      - Run with go run (development)"
	@echo "  test         - Run tests"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Build and run Docker container"
	@echo "  docker-stop  - Stop Docker container"
	@echo "  docker-rm    - Remove Docker container"
	@echo "  compose up   - Start with Docker Compose"
	@echo "  compose down - Stop Docker Compose services"
	@echo "  compose logs - Read Docker Compose logs"
	@echo "  tidy         - Clean up Go modules"
	@echo "  clean        - Clean build artifacts"