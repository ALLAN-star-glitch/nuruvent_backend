.PHONY: help migrate-up migrate-down migrate-status migrate-create migrate-reset dev build test lint worker worker-build seed seed-only seed-dry seed-reset db-shell db-list db-tables db-desc db-query db-count db-psql

# Migration directory
MIGRATION_DIR=internal/database/migrations

# Database URL - Using nuruvent_user (port 54582)
DB_HOST=localhost
DB_PORT=54582
DB_USER=nuruvent_user
DB_PASSWORD=nuruvent_user
DB_NAME=nuruvent
DB_SSL=disable
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)

# Docker container name
DB_CONTAINER=nuruvent-postgres

# ================================================
# HELP
# ================================================

help:
	@echo "Available commands:"
	@echo ""
	@echo "  Development:"
	@echo "  make dev                      - Run the API server"
	@echo "  make worker                   - Run the background worker"
	@echo "  make dev-all                  - Run both API and worker in parallel"
	@echo "  make build                    - Build the API binary"
	@echo "  make worker-build             - Build the worker binary"
	@echo "  make build-all                - Build both binaries"
	@echo "  make test                     - Run tests"
	@echo "  make test-coverage            - Run tests with coverage report"
	@echo "  make lint                     - Run linter"
	@echo "  make lint-fix                 - Auto-fix lint issues"
	@echo "  make clean                    - Clean build artifacts"
	@echo ""
	@echo "  Migrations:"
	@echo "  make migrate-create NAME=<name>  - Create a new migration"
	@echo "  make migrate-up                  - Apply all pending migrations"
	@echo "  make migrate-down                - Rollback the last migration"
	@echo "  make migrate-status              - Check migration status"
	@echo "  make migrate-reset               - Reset all migrations (dangerous!)"
	@echo ""
	@echo "  Seeders:"
	@echo "  make seed                     - Run all seeders"
	@echo "  make seed-only NAME=<seeder>  - Run only a specific seeder"
	@echo "  make seed-dry                 - Preview what would be seeded (dry run)"
	@echo "  make seed-reset               - Reset all seeders (dangerous!)"
	@echo "  make seed-force               - Force re-seed even if already run"
	@echo ""
	@echo "  Database (PostgreSQL):"
	@echo "  make db-shell                 - Open psql shell"
	@echo "  make db-list                  - List all databases"
	@echo "  make db-tables                - List all tables"
	@echo "  make db-desc TABLE=<name>     - Describe table structure"
	@echo "  make db-count TABLE=<name>    - Count rows in a table"
	@echo "  make db-query QUERY=<sql>     - Run a custom SQL query"
	@echo "  make db-psql CMD=<command>    - Run a psql command"
	@echo ""
	@echo "  Docker:"
	@echo "  make docker-up                - Start Docker services"
	@echo "  make docker-down              - Stop Docker services"
	@echo "  make docker-logs              - View Docker logs"
	@echo "  make docker-restart           - Restart Docker services"
	@echo ""
	@echo "  Swagger:"
	@echo "  make swagger                  - Generate Swagger documentation"
	@echo "  make swagger-fmt              - Format Swagger comments"
	@echo "  make swagger-install          - Install swag CLI"
	@echo "  make swagger-clean            - Remove generated docs"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate-create NAME=add_slug_to_business"
	@echo "  make db-desc TABLE=users"
	@echo "  make db-count TABLE=events"
	@echo "  make db-query QUERY=\"SELECT * FROM users LIMIT 5\""
	@echo "  make dev"
	@echo "  make worker"
	@echo "  make seed-only NAME=permissions"
	@echo "  make c 						-Clear terminal

# ================================================
# MIGRATIONS
# ================================================

# Sequential numbering for migrations
migrate-create:
	@echo "Creating migration: $(NAME)"
	@COUNT=$$(ls -1 $(MIGRATION_DIR)/*.sql 2>/dev/null | wc -l | tr -d ' '); \
	NEXT=$$(printf "%03d" $$((COUNT + 1))); \
	goose -dir $(MIGRATION_DIR) create $$NEXT"_$(NAME)" sql

migrate-up:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" down

migrate-status:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" status

migrate-reset:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" reset

# ================================================
# SEEDERS
# ================================================

# Run all seeders
seed:
	go run cmd/seed/main.go

# Run specific seeder
seed-only:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required"; \
		echo "Usage: make seed-only NAME=<seeder_name>"; \
		echo "Example: make seed-only NAME=permissions"; \
		exit 1; \
	fi
	go run cmd/seed/main.go -only=$(NAME)

# Dry run - preview what would be seeded
seed-dry:
	go run cmd/seed/main.go -dry-run

# Force re-seed (run even if already run)
seed-force:
	go run cmd/seed/main.go -force

# Reset all seeders (remove logs)
seed-reset:
	@echo "⚠️  This will remove all seeder logs, allowing seeders to run again."
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" != "yes" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	go run cmd/seed/main.go -force

# Seed in production (with confirmation)
seed-prod:
	go run cmd/seed/main.go -env=production

# Seed in production without confirmation (CI/CD)
seed-prod-ci:
	go run cmd/seed/main.go -env=production -skip-confirm

# ================================================
# DEVELOPMENT
# ================================================

# Run API server
dev:
	go run cmd/api/main.go

# Run worker
worker:
	go run cmd/worker/main.go

# Run both API and worker in parallel
dev-all:
	@echo "Starting API server and worker..."
	@make -j2 dev worker

# ================================================
# BUILD
# ================================================

# Build API binary
build:
	go build -o bin/nuruvent cmd/api/main.go

# Build worker binary
worker-build:
	go build -o bin/worker cmd/worker/main.go

# Build seed binary
seed-build:
	go build -o bin/seed cmd/seed/main.go

# Build all binaries
build-all: build worker-build seed-build
	@echo "Built API, worker, and seed binaries"

# ================================================
# TESTING & LINTING
# ================================================

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

# ================================================
# CLEANUP
# ================================================

clean:
	rm -rf bin/
	rm -f coverage.out
	rm -rf docs/

# ================================================
# DOCKER
# ================================================

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

# ================================================
# DATABASE (PostgreSQL)
# ================================================

# Open psql shell
db-shell:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

# List all databases
db-list:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\l"

# List all tables
db-tables:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

# Describe table structure
db-desc:
	@if [ -z "$(TABLE)" ]; then \
		echo "Error: TABLE is required"; \
		echo "Usage: make db-desc TABLE=<table_name>"; \
		echo "Example: make db-desc TABLE=users"; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\d $(TABLE)"

# Count rows in a table
db-count:
	@if [ -z "$(TABLE)" ]; then \
		echo "Error: TABLE is required"; \
		echo "Usage: make db-count TABLE=<table_name>"; \
		echo "Example: make db-count TABLE=users"; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT COUNT(*) FROM $(TABLE);"

# Run a custom SQL query
db-query:
	@if [ -z "$(QUERY)" ]; then \
		echo "Error: QUERY is required"; \
		echo "Usage: make db-query QUERY=\"<sql_query>\""; \
		echo "Example: make db-query QUERY=\"SELECT * FROM users LIMIT 5\""; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "$(QUERY)"

# Run a psql command
db-psql:
	@if [ -z "$(CMD)" ]; then \
		echo "Error: CMD is required"; \
		echo "Usage: make db-psql CMD=\"<command>\""; \
		echo "Example: make db-psql CMD=\"\dt\""; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "$(CMD)"

# Check database connection
db-ping:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT 1;"

# Show database size
db-size:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT pg_database_size('$(DB_NAME)');"

# Show all users
db-users:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\du"

# ================================================
# FULL SETUP
# ================================================

# Full setup: migrate + seed
setup:
	@echo "Running migrations..."
	@make migrate-up
	@echo "Running seeders..."
	@make seed
	@echo "✅ Setup complete!"

# Full setup with force re-seed
setup-force:
	@echo "Running migrations..."
	@make migrate-up
	@echo "Running seeders (force)..."
	@make seed-force
	@echo "✅ Setup complete!"

# ================================================
# SWAGGER
# ================================================

swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs

swagger-fmt:
	swag fmt

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

swagger-clean:
	rm -rf docs/

# ================================================
# WIRE - Dependency Injection
# ================================================

wire:
	@echo "Generating wire_gen.go..."
	wire gen ./internal/app

wire-install:
	go install github.com/google/wire/cmd/wire@latest

wire-verify:
	wire check ./...

wire-clean:
	rm -f internal/app/wire_gen.go

c:
	@echo "Clear Terminal"
	clear