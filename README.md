# KyupiKyupi Backend

KyupiKyupi is a matching application. This repository contains the Go service that handles:

- User registration and login with bcrypt-hashed passwords (users must be 18+ years old)
- JWT-based session management (24-hour access tokens)
- Authenticated profile retrieval and updates (name, gender, birth_date, bio, avatar_url, target_gender, intention)
- User logout (client-side token discard)
- Like and match functionality
- Real-time chat messaging between matched users
- Gin-powered HTTP server and routing

The data layer uses PostgreSQL for relational data (users, matches, likes) and MongoDB for chat messages. Docker Compose spins up both database systems.

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
- `POST /api/users` – create user account, returns JWT (requires user to be 18+ years old)
- `POST /api/users/sign_in` – authenticate, returns JWT

Protected endpoints (send `Authorization: Bearer <token>`):

- `GET /api/users/profile` – fetch current user profile
- `PATCH /api/users/profile` or `PUT /api/users/profile` – update profile fields
- `POST /api/users/profile/avatar` – upload or replace the avatar image (multipart/form-data with `avatar` field)
- `DELETE /api/users/sign_out` – sign out (client-side token discard)

Matching endpoints:

- `POST /api/v1/likes` – like or pass on a user
- `GET /api/v1/matches` – get all matches with user details

Chat endpoints (requires authentication and active match):

- `POST /api/v1/messages` – send a message to a matched user
- `GET /api/v1/matches/:match_id/messages` – get all messages for a specific match

Tokens expire after 24 hours by default. Logging out simply means discarding the token on the client.

### Field Specifications

**Gender Enums:**
Gender-related request/response payloads use integer enums:
- `1` = male
- `2` = female
- `3` = others

Both `gender` (required) and `target_gender` (optional) fields accept these values.

**Intention Values:**
The `intention` field (optional, defaults to "still_figuring_out") accepts one of:
- `long_term_partner` – Looking for a long-term partner
- `long_term_open_to_short` – Long-term relationship, open to short
- `short_term_open_to_long` – Short-term fun, open to long
- `short_term_fun` – Short-term fun
- `new_friends` – New friends
- `still_figuring_out` – Still figuring it out (default)

**Age Requirement:**
Users must be at least 18 years old to register. The `birth_date` field is validated during signup to ensure the user meets this requirement.

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
- `internal/db` – Database connectors (PostgreSQL and MongoDB)
- `internal/models` – domain models (e.g., `User`, `Match`, `Message`)
- `internal/repo` – data access layer for PostgreSQL and MongoDB
- `internal/auth` – JWT helper functions
- `internal/middleware` – authentication middleware for Gin
- `internal/handlers` – Gin handlers for auth/profile/health/match/chat endpoints
- `internal/routes` – Gin router wiring and middleware composition
- `docker-compose.yml` – local development stack
- `Makefile` – convenience tasks

## Chat Feature

The chat feature allows matched users to exchange messages:

- Messages are stored in MongoDB in the `messages` collection
- Each message contains: `match_id`, `sender_id`, `receiver_id`, `content`, and `created_at`
- Users can only send messages to users they have an active match with
- Users can only view messages for matches they are part of
- All chat endpoints require authentication via JWT token

---

## Contributing

Bug reports, feature ideas, and pull requests are welcome. For significant changes, open an issue or discussion first so we can coordinate design decisions.
