# Agents.md: Architectural Roles and Boundaries

This document defines the responsibilities for each module in the project. We follow **Clean Architecture** (Ports & Adapters) principles to ensure the business logic remains decoupled from external tools like Ory, Gin, or GORM.

---

## 1. Project Structure

```text
.
├── cmd/
│   └── api/                # Entry point: Cobra commands, config init, and server startup
├── internal/               # Private code
│   ├── domain/             # Entities (User, Stream) and Repository Interfaces
│   ├── service/            # Business Logic (Use Cases / Services)
│   ├── infrastructure/     # External implementations (Adapters)
│   │   ├── ory/            # Ory Kratos & Hydra SDK clients
│   │   ├── db/             # GORM & PostgreSQL implementation
│   │   ├── storage/        # MinIO & S3-compatible storage
│   │   └── mailer/         # SMTP & transactional email
│   ├── transport/          # Interaction protocols
│   │   └── http/           # Gin routes, controllers, and middleware
│   │       ├── handlers/
│   │       ├── middleware/
│   │       └── httputil/      # Common HTTP response helpers
├── pkg/                    # Public helper libraries
├── configs/                # YAML/Env configuration files
└── deployments/            # Docker Compose & K8s manifests
```

---

## 2. Layer Responsibilities

### A. CLI & Entry Point (Cobra)
**Location:** `cmd/api/`
- **Role:** Handles application startup, flag parsing, and dependency injection.
- **Rules:** - Initialize GORM, Ory Clients, and Services.
    - Setup the Gin engine and start the HTTP server.

### B. Transport Layer (Gin)
**Location:** `internal/transport/http/`
- **Middleware Agent:** - Intercepts requests to validate Bearer tokens via **Ory Hydra**.
    - Injects the `identity_id` into the `gin.Context`.
- **Handler Agent:** - Maps HTTP routes to **Service** methods.
    - Handles JSON binding and validation.
    - Returns standardized HTTP responses.

### C. Service Layer (Business Logic)
**Location:** `internal/service/`
- **Role:** The "Brain" of the application. Coordinates data flow between the Domain and Infrastructure.
- **Rules:** - Must NOT know about Gin, SQL, or Ory.
    - Works only with **Domain Entities** and **Interfaces**.
    - Implements core logic (e.g., "Can this user upload a video?").

### D. Domain Layer (Core)
**Location:** `internal/domain/`
- **Role:** Defines the "Language" of the project.
- **Rules:** - **Entities:** Simple Go structs representing objects (e.g., `User`, `Workspace`).
    - **NO SLOPPY TYPING:** Strictly **PROHIBITED** to use `map[string]interface{}` or `map[string]any` for representing Domain Entities or API payloads. Everything must be a named `struct`.
    - **Repository Interfaces:** Definitions of how data should be saved/retrieved (e.g., `type WorkspaceStore interface`).
    - No external dependencies allowed here.

### E. Infrastructure Layer (Adapters)
**Location:** `internal/infrastructure/`
- **Database Agent (GORM/Postgres):**
- **MIGRATIONS RULE:** Strictly **FORBIDDEN** to use `db.AutoMigrate()` inside the application startup flow.
- **MIGRATION TOOL:** All schema changes must be handled via explicit migration files/scripts (e.g., `gormigrate`).
- **SEPARATION:** Migrations must be triggered by a dedicated CLI command (e.g., `api migrate up`).
- **Ory Agent (Kratos):**
    - Wraps Ory SDKs to provide simplified methods for the application.
- **Storage Agent (MinIO):**
    - Handles file uploads, downloads, and lifecycle management.
    - Ensures required buckets (e.g., `studio`) exist upon initialization.
- **Mailer Agent (SMTP):**
    - Dispatches system emails, notifications, and auth-related correspondence.
    - Supports both standard SMTP and TLS-secured connections.

---

## 3. Communication Patterns

1. **Strong Typing Everywhere:** All data passing between layers (Transport -> Service -> Infrastructure) must use internal Domain structs.
2. **Dependency Injection:** All dependencies (DB, Ory, Services) must be passed via constructors (`New...`).
3. **Context Propagation:** Always pass `context.Context` from the Gin handler down to the GORM query.
4. **Ory Integration:** When interacting with Ory Kratos Admin API, use typed structs that mirror the `identity.schema.json`. Do not use raw maps for traits.

---

## 4. Technology Stack
- **CLI:** [Cobra](https://github.com/spf13/cobra)
- **Web Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **ORM:** [GORM](https://gorm.io/)
- **Auth:** [Ory Kratos](https://www.ory.sh/kratos/) (Identity)
- **Database:** PostgreSQL
- **Storage:** MinIO (S3-compatible)
- **Mailer:** SMTP (Maihog for dev)
