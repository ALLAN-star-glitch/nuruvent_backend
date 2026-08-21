.PHONY: help migrate-up migrate-down migrate-status migrate-create migrate-reset dev build test lint worker worker-build seed seed-only seed-dry seed-reset db-shell db-list db-tables db-desc db-query db-count db-psql db-drop db-reset cache-clear cache-clear-go cache-clear-all

# ================================================
# MIGRATION & DATABASE CONFIGURATION
# ================================================

# Migration directory
MIGRATION_DIR=internal/shared/database/migrations

# Database URL - Docker (uses postgres service name)
DB_HOST?=postgres
DB_PORT?=5432
DB_USER?=nuruvent_user
DB_PASSWORD?=nuruvent_user
DB_NAME?=nuruvent
DB_SSL?=disable
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)

# For local development (override)
LOCAL_DB_HOST?=localhost
LOCAL_DB_PORT?=54582
LOCAL_DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(LOCAL_DB_HOST):$(LOCAL_DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)

# Docker container name
DB_CONTAINER=nuruvent-postgres

# Go flags to ignore vendor directory
GOFLAGS=-mod=mod

# ================================================
# HELP
# ================================================

help:
	@echo "📋 Available commands:"
	@echo ""
	@echo "  🚀 Development:"
	@echo "  make dev                      - Run the API server"
	@echo "  make worker                   - Run the background worker"
	@echo "  make dev-all                  - Run both API and worker in parallel"
	@echo "  make build                    - Build the API binary"
	@echo "  make worker-build             - Build the worker binary"
	@echo "  make build-all                - Build all binaries"
	@echo "  make test                     - Run tests"
	@echo "  make test-coverage            - Run tests with coverage report"
	@echo "  make lint                     - Run linter"
	@echo "  make lint-fix                 - Auto-fix lint issues"
	@echo "  make clean                    - Clean build artifacts"
	@echo ""
	@echo "  🗄️  Migrations:"
	@echo "  make migrate-create NAME=<name>  - Create a new migration"
	@echo "  make migrate-up                  - Apply all pending migrations (Docker)"
	@echo "  make migrate-down                - Rollback the last migration (Docker)"
	@echo "  make migrate-status              - Check migration status (Docker)"
	@echo "  make migrate-up-local            - Apply migrations locally (localhost)"
	@echo "  make migrate-down-local          - Rollback migrations locally (localhost)"
	@echo "  make migrate-reset               - Reset ALL migrations (dangerous!)"
	@echo ""
	@echo "  🌱 Seeders:"
	@echo "  make seed                     - Run all seeders"
	@echo "  make seed-only NAME=<seeder>  - Run only a specific seeder"
	@echo "  make seed-dry                 - Preview what would be seeded (dry run)"
	@echo "  make seed-reset               - Reset all seeders (dangerous!)"
	@echo "  make seed-force               - Force re-seed even if already run"
	@echo ""
	@echo "  🐘 Database (PostgreSQL):"
	@echo "  make db-shell                 - Open psql shell"
	@echo "  make db-list                  - List all databases"
	@echo "  make db-tables                - List all tables"
	@echo "  make db-desc TABLE=<name>     - Describe table structure"
	@echo "  make db-count TABLE=<name>    - Count rows in a table"
	@echo "  make db-query QUERY=<sql>     - Run a custom SQL query"
	@echo "  make db-psql CMD=<command>    - Run a psql command"
	@echo "  make db-drop                  - ⚠️ DROP entire database (dangerous!)"
	@echo "  make db-reset                 - Drop and recreate database"
	@echo "  make db-backup                - Backup database to file"
	@echo "  make db-restore FILE=<name>   - Restore database from backup"
	@echo "  make db-seed-export           - Export seed data to SQL file"
	@echo ""
	@echo "  🧹 Cache & Cleanup:"
	@echo "  make cache-clear              - Clear all caches (Go, build, test)"
	@echo "  make cache-clear-go           - Clear only Go module cache"
	@echo "  make cache-clear-build        - Clear only build cache"
	@echo "  make cache-clear-test         - Clear only test cache"
	@echo "  make cache-clear-docker       - Clear Docker system cache"
	@echo "  make cache-clear-all          - Clear ALL caches (Go, build, test, Docker)"
	@echo "  make clean-all                - Clean everything (cache + build + wire)"
	@echo ""
	@echo "  🐳 Docker:"
	@echo "  make docker-up                - Start Docker services"
	@echo "  make docker-down              - Stop Docker services"
	@echo "  make docker-logs              - View Docker logs"
	@echo "  make docker-restart           - Restart Docker services"
	@echo "  make docker-rebuild           - Rebuild and restart Docker services"
	@echo "  make docker-clean             - Remove Docker containers and volumes"
	@echo ""
	@echo "  📚 Swagger:"
	@echo "  make swagger                  - Generate Swagger documentation"
	@echo "  make swagger-fmt              - Format Swagger comments"
	@echo "  make swagger-install          - Install swag CLI"
	@echo "  make swagger-clean            - Remove generated docs"
	@echo ""
	@echo "  🔧 Wire (DI):"
	@echo "  make wire                     - Generate wire_gen.go"
	@echo "  make wire-install             - Install wire CLI"
	@echo "  make wire-verify              - Check wire configuration"
	@echo "  make wire-clean               - Remove wire_gen.go files"
	@echo "  make wire-regen               - Clean and regenerate wire"
	@echo ""
	@echo "  🧹 Terminal:"
	@echo "  make c                        - Clear terminal"
	@echo ""
	@echo "📌 Examples:"
	@echo "  make migrate-create NAME=add_slug_to_business"
	@echo "  make db-desc TABLE=accounts"
	@echo "  make db-count TABLE=events"
	@echo "  make db-query QUERY=\"SELECT * FROM accounts LIMIT 5\""
	@echo "  make seed-only NAME=permissions"
	@echo "  make cache-clear-all"
	@echo "  make db-reset"

# ================================================
# MIGRATIONS
# ================================================

migrate-create:
	@echo "Creating migration: $(NAME)"
	@COUNT=$$(ls -1 $(MIGRATION_DIR)/*.sql 2>/dev/null | wc -l | tr -d ' '); \
	NEXT=$$(printf "%03d" $$((COUNT + 1))); \
	goose -dir $(MIGRATION_DIR) create $$NEXT"_$(NAME)" sql

migrate-up:
	@echo "🔄 Running migrations on Docker database..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" up

migrate-down:
	@echo "⬇️ Rolling back last migration on Docker database..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" down

migrate-status:
	@echo "📊 Migration status on Docker database..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" status

migrate-up-local:
	@echo "🔄 Running migrations on LOCAL database..."
	DB_HOST=localhost DB_PORT=54582 goose -dir $(MIGRATION_DIR) postgres "$(LOCAL_DB_URL)" up

migrate-down-local:
	@echo "⬇️ Rolling back last migration on LOCAL database..."
	DB_HOST=localhost DB_PORT=54582 goose -dir $(MIGRATION_DIR) postgres "$(LOCAL_DB_URL)" down

migrate-reset:
	@echo "⚠️  This will reset ALL migrations!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" != "yes" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" reset

# ================================================
# SEEDERS (with -mod=mod to ignore vendor)
# ================================================

seed:
	@echo "🌱 Running seeders..."
	go run $(GOFLAGS) cmd/seed/main.go

seed-only:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required"; \
		echo "Usage: make seed-only NAME=<seeder_name>"; \
		echo "Example: make seed-only NAME=permissions"; \
		exit 1; \
	fi
	@echo "🌱 Running seeder: $(NAME)"
	go run $(GOFLAGS) cmd/seed/main.go -only=$(NAME)

seed-dry:
	@echo "🌱 Previewing seeders (dry run)..."
	go run $(GOFLAGS) cmd/seed/main.go -dry-run

seed-force:
	@echo "🌱 Running seeders (force)..."
	go run $(GOFLAGS) cmd/seed/main.go -force

seed-reset:
	@echo "⚠️  This will remove all seeder logs, allowing seeders to run again."
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" != "yes" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	go run $(GOFLAGS) cmd/seed/main.go -force

seed-prod:
	@echo "🌱 Running seeders in production mode..."
	go run $(GOFLAGS) cmd/seed/main.go -env=production

seed-prod-ci:
	@echo "🌱 Running seeders in production mode (CI)..."
	go run $(GOFLAGS) cmd/seed/main.go -env=production -skip-confirm

# ================================================
# DEVELOPMENT
# ================================================

dev:
	go run $(GOFLAGS) cmd/api/main.go

worker:
	go run $(GOFLAGS) cmd/worker/main.go

dev-all:
	@echo "Starting API server and worker..."
	@make -j2 dev worker

# ================================================
# BUILD
# ================================================

build:
	go build -o bin/nuruvent cmd/api/main.go

worker-build:
	go build -o bin/worker cmd/worker/main.go

seed-build:
	go build -o bin/seed cmd/seed/main.go

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
# CACHE CLEANUP
# ================================================

cache-clear:
	@echo "🧹 Clearing all caches..."
	@go clean -modcache -cache -testcache
	@echo "✅ All caches cleared"

cache-clear-go:
	@echo "🧹 Clearing Go module cache..."
	@go clean -modcache
	@echo "✅ Go module cache cleared"

cache-clear-build:
	@echo "🧹 Clearing build cache..."
	@go clean -cache
	@echo "✅ Build cache cleared"

cache-clear-test:
	@echo "🧹 Clearing test cache..."
	@go clean -testcache
	@echo "✅ Test cache cleared"

cache-clear-docker:
	@echo "🧹 Clearing Docker system cache..."
	@docker system prune -f
	@echo "✅ Docker cache cleared"

cache-clear-all: cache-clear-go cache-clear-build cache-clear-test cache-clear-docker
	@echo "✅ All caches cleared"

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out
	@rm -rf docs/
	@echo "✅ Clean complete"

clean-all: cache-clear clean wire-clean
	@echo "🧹 Cleaned everything (cache, builds, wire_gen.go)"

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
	@rm -f internal/app/wire_gen.go
	@find . -name "wire_gen.go" -delete

wire-regen: wire-clean wire
	@echo "✅ Wire regenerated"

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

docker-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

docker-clean:
	@echo "⚠️  This will remove Docker containers, volumes, and images!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" != "yes" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	docker-compose down -v
	docker system prune -af
	@echo "✅ Docker cleaned"

# ================================================
# DATABASE (PostgreSQL)
# ================================================

db-shell:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

db-list:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\l"

db-tables:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

db-desc:
	@if [ -z "$(TABLE)" ]; then \
		echo "Error: TABLE is required"; \
		echo "Usage: make db-desc TABLE=<table_name>"; \
		echo "Example: make db-desc TABLE=accounts"; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\d $(TABLE)"

db-count:
	@if [ -z "$(TABLE)" ]; then \
		echo "Error: TABLE is required"; \
		echo "Usage: make db-count TABLE=<table_name>"; \
		echo "Example: make db-count TABLE=accounts"; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT COUNT(*) FROM $(TABLE);"

db-query:
	@if [ -z "$(QUERY)" ]; then \
		echo "Error: QUERY is required"; \
		echo "Usage: make db-query QUERY=\"<sql_query>\""; \
		echo "Example: make db-query QUERY=\"SELECT * FROM accounts LIMIT 5\""; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "$(QUERY)"

db-psql:
	@if [ -z "$(CMD)" ]; then \
		echo "Error: CMD is required"; \
		echo "Usage: make db-psql CMD=\"<command>\""; \
		echo "Example: make db-psql CMD=\"\dt\""; \
		exit 1; \
	fi
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "$(CMD)"

db-ping:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT 1;"

db-size:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT pg_database_size('$(DB_NAME)');"

db-users:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\du"

# ⚠️ DANGEROUS COMMANDS - Use with caution!
db-drop:
	@echo "⚠️  ⚠️  ⚠️  DANGEROUS COMMAND  ⚠️  ⚠️  ⚠️"
	@echo "This will DROP the ENTIRE database: $(DB_NAME)"
	@read -p "Are you ABSOLUTELY sure? Type the database name to confirm: " confirm; \
	if [ "$$confirm" != "$(DB_NAME)" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	@read -p "Type YES to confirm: " confirm2; \
	if [ "$$confirm2" != "YES" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	@echo "Terminating all connections to $(DB_NAME)..."
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_NAME)' AND pid <> pg_backend_pid();"
	@echo "Dropping database $(DB_NAME)..."
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "✅ Database $(DB_NAME) dropped"

db-reset: db-drop
	@echo "Recreating database $(DB_NAME)..."
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@echo "✅ Database $(DB_NAME) recreated"
	@echo "Running migrations..."
	@make migrate-up
	@echo "Running seeders..."
	@make seed
	@echo "✅ Database reset complete!"

db-backup:
	@echo "📦 Backing up database to backup_$(DB_NAME)_$(shell date +%Y%m%d_%H%M%S).sql..."
	@mkdir -p backups
	docker exec -it $(DB_CONTAINER) pg_dump -U $(DB_USER) $(DB_NAME) > backups/backup_$(DB_NAME)_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup complete"

db-restore:
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE is required"; \
		echo "Usage: make db-restore FILE=<backup_file.sql>"; \
		echo "Example: make db-restore FILE=backups/backup_nuruvent_20260811_123456.sql"; \
		exit 1; \
	fi
	@if [ ! -f "$(FILE)" ]; then \
		echo "Error: File $(FILE) not found"; \
		exit 1; \
	fi
	@echo "⚠️  Restoring database from $(FILE)"
	@read -p "This will overwrite current data. Continue? (yes/no): " confirm; \
	if [ "$$confirm" != "yes" ]; then \
		echo "Cancelled"; \
		exit 1; \
	fi
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) $(DB_NAME) < $(FILE)
	@echo "✅ Database restored from $(FILE)"

db-seed-export:
	@echo "📦 Exporting seed data to seed_data.sql..."
	@mkdir -p backups
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\copy (SELECT * FROM account_types ORDER BY slug) TO '/tmp/account_types.csv' WITH CSV HEADER"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\copy (SELECT * FROM professional_types ORDER BY slug) TO '/tmp/professional_types.csv' WITH CSV HEADER"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\copy (SELECT * FROM institution_types ORDER BY slug) TO '/tmp/institution_types.csv' WITH CSV HEADER"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\copy (SELECT * FROM event_types ORDER BY slug) TO '/tmp/event_types.csv' WITH CSV HEADER"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\copy (SELECT * FROM event_statuses ORDER BY slug) TO '/tmp/event_statuses.csv' WITH CSV HEADER"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\copy (SELECT * FROM media_types ORDER BY slug) TO '/tmp/media_types.csv' WITH CSV HEADER"
	@echo "✅ Seed data exported"

# ================================================
# FULL SETUP
# ================================================

setup:
	@echo "🚀 Running full setup..."
	@echo "Running migrations..."
	@make migrate-up
	@echo "Running seeders..."
	@make seed
	@echo "✅ Setup complete!"

setup-force:
	@echo "🚀 Running full setup (force re-seed)..."
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
# TERMINAL
# ================================================

c:
	@echo "Clear Terminal"
	clear