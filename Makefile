include .env
export

CONN_STRING = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR = ./internal/migrations

.PHONY: migration-create migration-up migration-down

migration-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Usage: make migration-create name=<migration_name>"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Run all pending migrations up
migration-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(CONN_STRING)" up

migration-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(CONN_STRING)" down

run-worker:
	go run ./cmd/worker