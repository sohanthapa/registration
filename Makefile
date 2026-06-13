APP_NAME=registration

.PHONY: run test fmt tidy docker-up docker-down migrate-up

run:
	go run ./cmd/app

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	psql "$${DATABASE_URL}" -f migrations/001_create_users.sql
