# Registration service

A simple registration service which supports:

- `POST /signup`
- `POST /login`
- password hashing with bcrypt
- JWT access token issuance
- PostgreSQL persistence

## Architecture

```text
cmd/app/main.go                 composition root / dependency wiring

internal/domain                 pure domain models and domain errors
internal/app                    use cases and ports/interfaces
internal/adapters/postgres      Postgres adapter for user persistence
internal/adapters/security      bcrypt and JWT adapters
internal/transport/http         HTTP handlers and request/response mapping
```

The application service owns the sign-up and login flows. HTTP, Postgres, bcrypt, and JWT are adapters around it.

## Requirements

- Go 1.22+
- Docker, optional but recommended for local Postgres
- PostgreSQL, if not using Docker

## Setup

Load the values into your shell:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable'
export JWT_SECRET='replace-this-with-a-long-random-secret'
export ADDR=':8080'
```

Start Postgres with Docker:

```bash
docker compose up -d
```

The migration runs automatically when the Docker volume is first created because the SQL file is mounted into `/docker-entrypoint-initdb.d`.

If you are using an existing database, run the migration manually:

```bash
psql "$DATABASE_URL" -f internal/adapters/postgres/migrations/001_create_users.sql
```

Install dependencies:

```bash
go mod tidy
```

Run the API:

```bash
go run ./cmd/app
```

Or use Make:

```bash
make docker-up
make tidy
make run
```

## Endpoints

### Health check

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "status": "ok"
}
```

### Sign up

```bash
curl -X POST http://localhost:8080/signup \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "supersecret123"
  }'
```

Response:

```json
{
  "user": {
    "id": "example-user-id",
    "email": "user@example.com",
    "created_at": "2026-06-11T12:00:00Z"
  }
}
```

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "supersecret123"
  }'
```

Response:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "expires_at": "2026-06-12T12:00:00Z",
  "user": {
    "id": "example-user-id",
    "email": "user@example.com",
    "created_at": "2026-06-11T12:00:00Z"
  }
}
```

## AI docs

Plans and logs from AI-assisted work can be found under `docs/cursorlog`.

## Production Deployment Guide

Guidelines to follow if we want to productionalize this service can be found in docs/productionalized.md file.

