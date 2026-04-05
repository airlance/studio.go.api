# Agents.md: Architectural Roles and Boundaries

This document defines the responsibilities for each module in the project. We follow **Clean Architecture** (Ports & Adapters) principles to ensure the business logic remains decoupled from external tools like Ory, Gin, or GORM.

---

## 1. Project Structure

```text
.
├── cmd/
│   ├── root.go             # Cobra root command, WaitGroup injection
│   ├── serve.go            # HTTP server startup and route wiring
│   └── migrate.go          # Migration CLI commands (up / down)
├── internal/               # Private code
│   ├── domain/             # Entities, Repository Interfaces, Service Interfaces
│   │   ├── mailer.go       # Mailer port (MailMessage, Mailer interface)
│   │   ├── profile.go
│   │   ├── user.go
│   │   └── workspace.go
│   ├── service/            # Business Logic (Use Cases)
│   │   ├── profile_service.go
│   │   └── workspace_service.go
│   ├── infrastructure/     # External adapters
│   │   ├── db/             # GORM / PostgreSQL
│   │   │   ├── migrations/
│   │   │   ├── profile_repository.go
│   │   │   └── workspace_repository.go
│   │   ├── mailer/         # Email delivery
│   │   │   └── mailer.go   # SMTP adapter (Mailhog + real providers)
│   │   ├── ory/            # Ory Kratos identity adapter
│   │   ├── rabbitmq/       # AMQP client
│   │   └── storage/        # MinIO / S3-compatible storage
│   ├── di/
│   │   └── container.go    # Dependency injection container
│   ├── config/
│   │   └── config.go       # Env-based config (envconfig + godotenv)
│   └── transport/
│       └── http/           # Gin routes, handlers, middleware
│           ├── handlers/
│           ├── middleware/
│           └── utils/      # CORS, respond helpers, error types
├── pkg/                    # Public helper libraries
├── configs/                # YAML / Env configuration files
└── deployments/            # Docker Compose & K8s manifests
```

---

## 2. Layer Responsibilities

### A. CLI & Entry Point (Cobra)
**Location:** `cmd/`
- **Role:** Application startup, flag parsing, and dependency injection.
- **Rules:**
    - `serve.go` — initialises `di.Container`, wires repositories → services → handlers, starts Gin.
    - `migrate.go` — opens a raw GORM connection and runs `gormigrate` commands; never shares the serve-time container.
    - No business logic here — delegate everything to the Service layer.

### B. Transport Layer (Gin)
**Location:** `internal/transport/http/`
- **Middleware Agent:**
    - `middleware/auth.go` — validates `ory_kratos_session` cookie via Kratos FrontendAPI.
    - Injects `*ory.Identity` into `gin.Context` under the key `"user"`.
    - Rejects unverified identities with `403 Forbidden`.
- **Handler Agent:**
    - Maps HTTP routes to **Service** methods.
    - Handles multipart form binding and JSON binding.
    - Returns standardised responses via `utils.RespondOK` / `utils.RespondError`.
- **Utils:**
    - `utils/cors.go` — origin-whitelist CORS middleware.
    - `utils/respond.go` / `utils/response.go` — typed response helpers.

### C. Service Layer (Business Logic)
**Location:** `internal/service/`
- **Role:** The "Brain" of the application. Coordinates data flow between Domain and Infrastructure.
- **Rules:**
    - Must NOT know about Gin, SQL, Ory SDKs, or SMTP internals.
    - Works only with **Domain Entities** and **Interfaces** (`domain.ProfileRepository`, `domain.Storage`, `domain.Mailer`, …).
    - Implements core logic (e.g., "Can this user upload a video?", "Send welcome email after workspace creation").
    - Receives `domain.Mailer` via constructor injection when email dispatch is needed.

### D. Domain Layer (Core)
**Location:** `internal/domain/`
- **Role:** Defines the "Language" of the project.
- **Rules:**
    - **Entities:** Simple Go structs (`Profile`, `Workspace`, `WorkspaceMember`, `WorkspaceInvite`, `UserWorkspaceConfig`, `User`).
    - **NO SLOPPY TYPING:** Strictly **PROHIBITED** to use `map[string]interface{}` or `map[string]any` for Domain Entities or API payloads. Everything must be a named `struct`.
    - **Repository Interfaces:** `ProfileRepository`, `WorkspaceRepository`, `UserRepository`.
    - **Service Interfaces:** `ProfileService`, `WorkspaceService`.
    - **Port Interfaces:** `Storage`, `Mailer` — infrastructure ports consumed by services.
    - No external dependencies allowed here.

#### Mailer Port (`domain/mailer.go`)
```go
type MailMessage struct {
    To      []string
    Subject string
    HTML    string   // rendered HTML body
    Text    string   // optional plain-text fallback
    ReplyTo string
    CC      []string
}

type Mailer interface {
    Send(ctx context.Context, msg MailMessage) error
}
```

### E. Infrastructure Layer (Adapters)
**Location:** `internal/infrastructure/`

#### Database Agent (`db/`)
- Implements `ProfileRepository` and `WorkspaceRepository` using GORM.
- **MIGRATIONS RULE:** Strictly **FORBIDDEN** to use `db.AutoMigrate()` inside the application startup flow.
- **MIGRATION TOOL:** All schema changes live in `db/migrations/` as explicit `gormigrate` entries.
- **SEPARATION:** Migrations are triggered only by `api migrate up` / `api migrate down`.

#### Mailer Agent (`mailer/`)
- Single file `mailer.go` — one SMTP adapter, one constructor: `New(*config.MailerConfig) (domain.Mailer, error)`.
- **Behaviour adapts automatically to config — no provider switch needed:**
    - `MAILER_USERNAME` empty → no auth (Mailhog, local dev).
    - `MAILER_USERNAME` set → `smtp.PlainAuth` (real provider with STARTTLS).
    - `MAILER_PORT=465` → implicit TLS via `crypto/tls` (e.g. Gmail SMTP, Brevo).
    - Any other port → `net/smtp.SendMail` with optional auth (STARTTLS negotiated by server).
- Assembles `multipart/alternative` MIME (HTML + auto-generated plain-text fallback).
- **To switch from Mailhog to a real provider** — update `.env` only; no code changes required.

#### Ory Agent (`ory/`)
- Wraps the Ory Kratos Admin API SDK.
- Implements `domain.UserRepository` (`FindByID`, `GetIdentity`).
- Uses typed structs mirroring `identity.schema.json` — never raw maps for traits.

#### Storage Agent (`storage/`)
- Implements `domain.Storage` via MinIO client.
- Manages buckets (`workspaces`, `profiles`): creates them and sets public-read policy on startup.
- `GetPresignedURL` returns a public URL using `STORAGE_PUBLIC_BASE_URL` (configurable for local vs. cloud).

#### RabbitMQ Agent (`rabbitmq/`)
- Optional dependency — if the broker is unavailable at startup, the container degrades gracefully (no panic).
- Exposes `Publish` and `DeclareQueue` behind a mutex-protected `amqp.Channel`.

---

## 3. Dependency Injection (`di/Container`)

`Container` is the single composition root. It holds:

| Field      | Type              | Source                          |
|------------|-------------------|---------------------------------|
| `Config`   | `*config.Config`  | `config.Init(ctx)`              |
| `DB`       | `*gorm.DB`        | `postgres.Open(cfg.DB.DSN)`     |
| `Storage`  | `domain.Storage`  | `storage.NewMinioStorage(cfg)`  |
| `Mailer`   | `domain.Mailer`   | `mailer.New(&cfg.Mailer)`       |
| `RabbitMQ` | `*rabbitmq.Client`| optional, degrades gracefully   |

Services receive only the interfaces they need — never the full container.

---

## 4. Communication Patterns

1. **Strong Typing Everywhere:** All data passing between layers (Transport → Service → Infrastructure) uses internal Domain structs.
2. **Dependency Injection:** All dependencies (DB, Storage, Mailer, …) are passed via `New…` constructors.
3. **Context Propagation:** `context.Context` is always threaded from the Gin handler down to GORM queries and `Mailer.Send`.
4. **Ory Integration:** Kratos Admin API is accessed via typed SDK structs — no raw maps for identity traits.
5. **Graceful Degradation:** Optional infrastructure (RabbitMQ) must not prevent startup on failure.

---

## 5. Configuration Reference

Relevant env vars for the Mailer (see `.env` and `config/config.go`):

| Variable          | Default                     | Description                                     |
|-------------------|-----------------------------|--------------------------------------------------|
| `MAILER_FROM`     | `no-reply@studio.localhost` | Envelope sender address                          |
| `MAILER_HOST`     | `localhost`                 | SMTP server hostname                             |
| `MAILER_PORT`     | `1025`                      | `1025` = Mailhog, `587` = STARTTLS, `465` = TLS |
| `MAILER_USERNAME` | —                           | Leave empty for Mailhog; set for real providers  |
| `MAILER_PASSWORD` | —                           | SMTP auth password                               |

---

## 6. Technology Stack

| Concern       | Tool                                                                 |
|---------------|----------------------------------------------------------------------|
| CLI           | [Cobra](https://github.com/spf13/cobra)                             |
| Web Framework | [Gin Gonic](https://github.com/gin-gonic/gin)                       |
| ORM           | [GORM](https://gorm.io/)                                            |
| Migrations    | [gormigrate](https://github.com/go-gormigrate/gormigrate)           |
| Auth          | [Ory Kratos](https://www.ory.sh/kratos/) (Identity & Session)       |
| Database      | PostgreSQL                                                           |
| Storage       | MinIO (S3-compatible)                                               |
| Mailer        | SMTP (`net/smtp` + `crypto/tls`) / Log (dev)                        |
| Message Queue | RabbitMQ via [amqp091-go](https://github.com/rabbitmq/amqp091-go)  |