# kyupi-kyupi-backend

Minimal Go web backend scaffold.

Quick start

1. Build:

  go build ./...

2. Run:

  go run ./

The server listens on :8080 and exposes:

- GET / -> {"message": "kyupi-kyupi-backend"}
- GET /health -> {"status": "ok"}

Run tests:

   go test ./...

Environment

This project reads configuration from environment variables. You can copy `.env.example` to `.env` for local development or export variables directly.

Important variables:

- `APP_ENV` — one of `DEVELOPMENT`, `TESTING`, `PRODUCTION`. Defaults to `DEVELOPMENT`.
- `PORT` — TCP port the server listens on (default `8080`).
- `LOG_LEVEL` — `debug`, `info`, `warn`, `error` (default `info`).
- `DATABASE_URL` — optional database connection string.

The project includes an `internal/config` package that loads these values and provides helpers such as `Addr()` and `IsTesting()`.

CI

A GitHub Actions workflow is provided at `.github/workflows/ci.yml` and will run `go test` and `go build` on push/PR targeting `main`.
