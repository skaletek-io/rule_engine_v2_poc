BINARY_NAME=rule-engine-v2-poc

MIGRATIONS_PATH = internal/platform/db/migrations/sql

DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=skl_monitoring_v2_db

deps:
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

generate:
	oapi-codegen --config=api/oapi-codegen.yaml api/openapi.yaml
	sqlc generate

build:
	go build -o bin/$(BINARY_NAME) ./cmd/api/main.go

run:
	go run ./cmd/api/main.go

test:
	go test ./...

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

migrate-reset:
	migrate -path $(MIGRATIONS_PATH) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" reset