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
# kyupi-kyupi-backend