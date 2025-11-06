# KyupiKyupi Backend

KyupiKyupi is a matching application. This repository contains the Go service that handles:

- User registration and login with bcrypt-hashed passwords
- JWT-based session management (24-hour access tokens)
- Authenticated profile retrieval and updates (name, gender, birth_date, bio, avatar_url)
- User logout (client-side token discard)
- Gin-powered HTTP server and routing

The data layer is PostgreSQL, accessed through `database/sql`. Docker Compose spins up PostgreSQL (and a MongoDB container reserved for future features).

---

## Requirements

Prepare the following before running KyupiKyupi locally:

- Go 1.21+
- PostgreSQL 15
- MongoDB 6.0
- Docker 24+ with Docker Compose v2 (recommended for local development)
- Make (optional, offers build/test shortcuts)

---

## Configuration

Copy the sample environment file and adjust values as needed:

```bash
cp .env.example .env
```

Key environment variables:

- `APP_ENV` – `DEVELOPMENT`, `TESTING`, or `PRODUCTION` (default `DEVELOPMENT`)
- `PORT` – HTTP port (default `8080`)
- `LOG_LEVEL` – `debug`, `info`, `warn`, or `error`
- `LOG_DIR` – directory where runtime logs are stored (default `log`)
- `JWT_SECRET` – required; used to sign access tokens (change this in production)
- PostgreSQL configuration (required)
   - `POSTGRES_URL`, or
   - `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`
   - Pooling & SSL: `POSTGRES_SSL_MODE` (default `disable`), `POSTGRES_MAX_OPEN_CONNS`, `POSTGRES_MAX_IDLE_CONNS`, `POSTGRES_CONN_MAX_LIFETIME` (Go duration format, default `30m`)
- MongoDB configuration (required)
   - `MONGODB_URL`, or
   - `MONGO_HOST`, `MONGO_PORT`, `MONGO_USER`, `MONGO_PASSWORD`, `MONGO_DATABASE`, `MONGO_AUTH_SOURCE`, `MONGO_REPLICA_SET`
- Avatar storage configuration (optional; defaults shown)
   - `AVATAR_STORAGE_DIR` – filesystem path for uploaded avatars (default `storage/avatars`)
   - `AVATAR_URL_PREFIX` – public URL prefix served by the API (default `/avatars`)

At startup the service ensures the `users` table exists in PostgreSQL, so you do not need a separate migration step for basic usage.

---

## Run With Docker (recommended)

1. Build and start the stack (API + PostgreSQL + MongoDB placeholder):

   ```bash
   docker compose up -d
   ```

2. Follow logs:

   ```bash
   docker compose logs -f app
   ```

3. Access the API at <http://localhost:8080>.

4. Shut everything down:

   ```bash
   docker compose down
   ```

Persistent volumes `pgdata` and `mongodata` store database data between restarts. Avatars persist in `./storage/avatars` on the host (mounted into the container). Use `docker compose down -v` to remove database volumes, and delete the `storage/` directory if you want to clear uploaded avatars. The API validates connections to both PostgreSQL and MongoDB during startup.

---

## API Documentation (Swagger)

- When the server is running, visit <http://localhost:8080/swagger/index.html> to view the interactive docs.
- The generated specification is committed under `docs/` (`swagger.json` / `swagger.yaml`).
- To regenerate after changing handler annotations:

   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   $(go env GOPATH)/bin/swag init --output docs --parseInternal --parseDependency
   ```

   The CLI is only required when documentation changes are made.

---

## Run Without Docker

1. Preparation: Install Go, PostgreSQL and MongoDB

2. Start PostgreSQL and MongoDB services, then create/verify the databases (defaults match `.env.example`). Examples:

   ```bash
   psql -h localhost -U postgres -c 'CREATE DATABASE kyupi;'
   mongosh --eval 'db.getSiblingDB("kyupi").runCommand({ ping: 1 })'
   ```

3. Export environment variables or rely on `.env` (with tools like `direnv`).

4. Install Go dependencies and start the server (startup pings both PostgreSQL and MongoDB; ensure they are reachable):

   ```bash
   go mod tidy
   go run .
   ```

   The service listens on `PORT`, defaulting to `:8080`.

4. Optional Make targets:

   ```bash
   make build   # go build -v ./...
   make run     # go run ./
   make test    # go test ./... -v
   make tidy    # go mod tidy
   ```

---

## API Summary

Public endpoints:

- `GET /` – welcome payload
- `GET /health` – readiness probe
- `POST /api/users` – create user account, returns JWT
- `POST /api/users/sign_in` – authenticate, returns JWT

Protected endpoints (send `Authorization: Bearer <token>`):

- `GET /api/users/profile` – fetch current user profile
- `PATCH /api/users/profile` or `PUT /api/users/profile` – update profile fields
- `POST /api/users/profile/avatar` – upload or replace the avatar image (multipart/form-data with `avatar` field)
- `DELETE /api/users/sign_out` – sign out (client-side token discard)

Tokens expire after 24 hours by default. Logging out simply means discarding the token on the client.

---

## Testing

```bash
go test ./...
```

GitHub Actions (`.github/workflows/ci.yml`) runs `go mod tidy`, `go test ./...`, and `go build ./...` on pushes and pull requests targeting `main`.

---

## Project Structure

- `main.go` – application bootstrap, HTTP server, DB setup
- `internal/config` – environment-driven settings
- `internal/db` – Database connector
- `internal/models` – domain models (e.g., `User`)
- `internal/repo` – data access layer
- `internal/auth` – JWT helper functions
- `internal/middleware` – authentication middleware for Gin
- `internal/handlers` – Gin handlers for auth/profile/health endpoints
- `internal/routes` – Gin router wiring and middleware composition
- `docker-compose.yml` – local development stack
- `Makefile` – convenience tasks

---

## Contributing

Bug reports, feature ideas, and pull requests are welcome. For significant changes, open an issue or discussion first so we can coordinate design decisions. 
Test push
