# AGENTS.md

## Project Overview

REST API service built with Go. Uses Cobra for CLI entrypoint and command
management, Logrus for structured logging, and Gin for HTTP routing and
middleware.

## Repository Layout
```
.
├── cmd/
│   ├── root.go          # Root cobra command, global flags, logger init
│   └── serve.go         # `serve` subcommand — starts HTTP server
├── internal/
│   ├── config/          # Config struct + loader (env/file)
│   ├── handler/         # Gin handler functions, grouped by domain
│   ├── middleware/       # Custom Gin middleware (auth, logging, recovery)
│   ├── service/         # Business logic, no HTTP/transport concerns
│   └── repository/      # Data access layer
├── pkg/                 # Reusable packages safe to import externally
├── api/                 # OpenAPI/Swagger specs
├── migrations/          # DB migration files
├── config.yaml          # Default config
├── main.go              # Entry: calls cmd.Execute()
└── Makefile
```

## Tech Stack

| Package | Role |
|---|---|
| `github.com/spf13/cobra` | CLI commands and flags |
| `github.com/sirupsen/logrus` | Structured JSON logging |
| `github.com/gin-gonic/gin` | HTTP router, middleware, binding |

## Commands
```bash
make run          # go run ./main.go serve
make build        # go build -o bin/api ./main.go
make test         # go test ./...
make lint         # golangci-lint run
make migrate-up   # run DB migrations
```

The primary runtime command is:
```bash
./api serve --port 8080 --config config.yaml
```

## Cobra Conventions

- `cmd/root.go` initialises the Logrus logger and binds global flags
  (`--log-level`, `--config`). All subcommands inherit these.
- Each subcommand lives in its own file under `cmd/`.
- Config is loaded inside `PersistentPreRunE` on the root command so every
  subcommand has access before `Run` executes.
- Never call `os.Exit` directly — return errors up to `cobra.Command.RunE`
  and let Cobra handle exit codes.

## Logrus Conventions

- Logger is initialised once in `cmd/root.go` and passed via context or
  injected into structs — never use the global `logrus.Info(...)` calls
  outside of `cmd/`.
- Log format: `logrus.JSONFormatter` in production, `TextFormatter` locally
  (driven by `LOG_FORMAT` env var).
- Always log with fields, not interpolated strings:
```go
  // good
  log.WithFields(logrus.Fields{"user_id": id, "op": "GetUser"}).Info("fetching user")

  // bad
  log.Infof("fetching user %d", id)
```
- Standard field keys: `op`, `user_id`, `request_id`, `duration_ms`, `error`.
- Request ID is injected by the `middleware.RequestID` middleware and stored
  in Gin context. Handlers retrieve it with `middleware.GetRequestID(c)`.

## Gin Conventions

- Router is constructed in `internal/handler/router.go` via `NewRouter(deps)`.
  Do not register routes in `cmd/`.
- Group routes by domain:
```go
  v1 := r.Group("/api/v1")
  {
      users := v1.Group("/users")
      users.GET("/:id", h.User.Get)
      users.POST("", h.User.Create)
  }
```
- Always use `c.ShouldBindJSON(&req)` and validate with struct tags. Return
  `400` with a JSON error body on bind failure.
- Standard JSON error response shape:
```json
  { "error": "message", "code": "SNAKE_CASE_CODE" }
```
- Middleware stack (applied in order): `Recovery` → `RequestID` →
  `Logger` → `Auth` (on protected groups).
- The `Logger` middleware logs method, path, status, latency, and
  `request_id` via Logrus at `Info` level. Errors are logged at `Error`.

## Error Handling

- Services return typed errors from `internal/apierr` package.
- Handlers translate service errors to HTTP status codes via a shared
  `handler.RespondError(c, err)` helper — no raw `c.JSON(500, ...)` calls.
- Never swallow errors. Every `err != nil` must either be returned or logged.

## Testing

- Unit tests live alongside source files (`foo_test.go`).
- Integration tests under `internal/handler/` use `httptest` + real Gin
  router with a stub service layer.
- Use `testify/assert` and `testify/require`. `require` for setup assertions
  that would make the rest of the test meaningless if they fail.
- Mock interfaces with `mockery` — generated mocks in `internal/mocks/`.

## Configuration

All config is loaded from `config.yaml` with env var overrides
(`APP_PORT`, `APP_DB_DSN`, etc.). The `internal/config` package exposes a
`Config` struct. No `os.Getenv` calls outside of that package.

## Code Style

- `gofmt` + `goimports` enforced by CI.
- Public functions on handlers/services must have a godoc comment.
- Keep handler functions thin: bind → call service → respond. Business logic
  belongs in `service/`.
- Avoid `init()` functions. Explicit wiring in `main.go` and `cmd/` only.