# DeepLab Task System – Folder Usage

This monorepo powers the multi-tenant remote lab platform described as “多用户远程开发 + 资源隔离平台（轻量版）” with queued, prioritized DeepLearning tasks. Use this guide to quickly locate the code that matters when building agents, services, or tooling on top of the platform.

## High-Level Breakdown

| Path | What Lives Here | Notes for Agents |
| --- | --- | --- |
| `api/` | Versioned OpenAPI/Thrift/etc. interface specs (currently `v1/`). | Source of truth for RPC/REST contracts; keep agents aligned with the spec before touching services. |
| `cmd/platform`, `cmd/scheduler`, `cmd/worker` | Go entrypoints for the three services (user/platform API, task scheduler, execution worker). | Each main links to the corresponding `internal/<service>` package. Extend or add binaries here. |
| `internal/` | Service implementations and shared libraries. | See the section below for a deeper map. |
| `web/` | React + TypeScript + Vite front-end for the lab control panel. | Hooks into the `platform` service; uses the standard Vite stack (`App.tsx`, `main.tsx`, global styles). |
| `configs/` | Environment-specific TOML/YAML settings (e.g., `conf/app.toml` read by the platform server). | Store per-env overrides here; never hardcode secrets in source. |
| `deployments/` | IaC manifests (Docker Compose, Helm, K8s) for promoting the cluster. | Empty scaffold today—add manifests as the platform matures. |
| `docs/` | Living documentation for contributors (this file belongs here). | Add design notes, onboarding guides, API references, etc. |
| `scripts/docker/` | Utility scripts, currently `pg.sh` for spinning up a local Postgres 18 instance with sane defaults. | Agents can rely on this for reproducible dev databases. |

## Internal Packages

| Path | Purpose | Key Technologies |
| --- | --- | --- |
| `internal/lib/apis` | Shared API helpers (request builders, bindings). | Go |
| `internal/lib/auth` | Authentication/session helpers for multi-user isolation. | Go |
| `internal/lib/config` | Centralized config loading/parsing. | Go + `github.com/BurntSushi/toml` |
| `internal/lib/db` | Database models such as `Container` (GORM structs) reused across services. | Go + GORM |
| `internal/lib/docker` | Docker client manager for container lifecycle + SSH port binding. | Go + Docker SDK |
| `internal/lib/errors` | Error definitions/mappers. | Go |
| `internal/lib/logger` | Zap + Lumberjack logging setup (file rotation for info/warn/error tiers). | `go.uber.org/zap`, `lumberjack` |
| `internal/lib/middleware` | HTTP middleware (sessions, auth, tracing) shared by Gin servers. | `github.com/gin-gonic/gin` |
| `internal/lib/models` | Higher-level domain models beyond raw DB structs. | Go |
| `internal/lib/queue` | Definitions for task queue payloads and priorities. | Go |
| `internal/platform/{handlers,repository,router,services,singleton,types}` | User/platform API layer: HTTP handlers, data access, service logic, dependency singletons, DTOs. | Gin, GORM, Zap |
| `internal/scheduler/{cron,dispatcher,handlers,repository,router,services,types}` | Task scheduling brain: cron definitions, dispatch logic, and API for manipulating queue priorities and preemption. | Go routines, cron jobs |
| `internal/worker/{consumer,db,grpc,handlers,processors}` | Execution worker that consumes queued jobs, talks to DB/GRPC backends, and runs containerized DL tasks. | Go + gRPC |

## How the Pieces Fit

1. **Platform service** (`cmd/platform` + `internal/platform`): exposes HTTP APIs (Gin) for workspace lifecycle, user management, and now Docker-backed container provisioning (see `/api/v1/containers`). Sessions use cookie-backed storage per `platform/router`.
2. **Scheduler** (`cmd/scheduler` + `internal/scheduler`): reads task intents, applies priority/queue rules, and dispatches runnable jobs to workers or external queues.
3. **Worker** (`cmd/worker` + `internal/worker`): consumes tasks, provisions isolated containers (see `internal/lib/db.Container`), executes workloads, and reports status.
4. **Web UI** (`web/`): React/Vite dashboard where lab users manage submissions and monitor queue states. Talks to the platform API.
5. **Supporting assets** (`scripts/docker/pg.sh`, `configs/`, `deployments/`): help agents bootstrap local dependencies (Postgres) and describe runtime environments.

Keep this map handy when extending multi-user features (access control in `internal/lib/auth`), improving isolation (worker/processors), or evolving scheduling policies (`internal/scheduler/dispatcher`). Update this file whenever new directories appear so fellow agents can onboard instantly.

## Layering Rules

- `internal/lib` is the shared foundation for cross-service helpers, adapters, and data contracts. Keep it limited to reusable primitives—never place platform/scheduler/worker business logic here.
- `internal/platform`, `internal/scheduler`, and `internal/worker` must remain fully isolated. They may import packages from `internal/lib` but must not import each other directly or indirectly.
- `cmd/<service>` entrypoints must only wire together their respective service package(s) plus `internal/lib` dependencies. Avoid side effects that would introduce hidden coupling between services.
- When adding new shared capabilities, factor them into `internal/lib/<area>` (or a new subpackage) so that each service consumes the same tested implementation rather than duplicating code.
- Enforce the layering with tooling—e.g., static analysis or CI checks that forbid intra-service imports—to keep the boundaries solid as the monorepo grows.
