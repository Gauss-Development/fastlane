# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Fiberlane — Photonics Sourcing MVP

3-day MVP for an AI-powered B2B sourcing platform connecting US engineers with verified Chinese photonics suppliers. Cross-border (Shenzhen → San Francisco) industrial design, hybrid AI search, magic-link supplier RFQ flow.

- **Vertical (first wedge)**: fiber-optic transceivers — SFP / SFP+ / QSFP / QSFP28.
- **Linear**: project `Fiberlane — Photonics Sourcing MVP` (id `4d710fa0-1c19-4d9b-82da-33ab683ebc2b`), team `Gaussdev` (`GAU`). All work tracked there. Milestones: Day 1 Foundation, Day 2 AI Search & Buyer Flow, Day 3 Polish/Landing/Demo.
- **Reference docs in Linear** (read these before changing the relevant area):
  - [Screen Inventory — 14 Screens, Build Order, Demo Purpose](https://linear.app/gaussdev/document/screen-inventory-14-screens-build-order-demo-purpose-5cd691675118)
  - [Design System — Industrial, Cross-Border, No AI-Slop](https://linear.app/gaussdev/document/design-system-industrial-cross-border-no-ai-slop-a85d2f22ac68)
  - [Order Service Architecture](https://linear.app/gaussdev/document/order-service-architecture-3344e8fea2be)

This codebase is **forked from a microblog gRPC template**. Several services from the template are being repurposed or replaced. See "Template inheritance" below before assuming a service belongs to Fiberlane.

## Repository layout

Polyglot monorepo: Go microservices + Next.js frontend.

- `proto/` — shared gRPC contracts as its own Go module (`github.com/nikitashilov/microblog_grpc/proto`). Each domain lives under `proto/<domain>/v1/`. Consumed by every service via `replace github.com/nikitashilov/microblog_grpc/proto => ../../proto` in their `go.mod`. Generated `.pb.go` / `_grpc.pb.go` are committed.
- `services/<name>/` — Go modules, one per service. Each has its own `go.mod`, `Dockerfile` (built from repo root so `COPY proto/` works), `main.go`.
- `frontend/` — Next.js 16 App Router app (React 19, Tailwind 4, TanStack Query, Zustand). Package manager is **bun** (`bun.lock`). BFF routes live under `frontend/src/app/api/bff/*` and proxy to the api-gateway.
- `docker-compose.yml` + `Makefile` — full-stack orchestration. Only `api-gateway`, Prometheus, and Grafana are exposed to the host; everything else is on internal Docker networks (`internal_net`, `edge_net`). The gateway's host port comes from `API_GATEWAY_PORT` (default `8080`). **Local override**: this machine has another project on `:8080`, so `.env` uses `API_GATEWAY_PORT=8090` — if you change it, also set `BACKEND_API_URL` in `frontend/.env.local` to match.
- `monitoring/` — Prometheus scrape config + Grafana provisioning. See `docs/monitoring.md`.
- `scripts/postgres-init-*.sql` — bootstrap SQL mounted into per-service Postgres containers at first start.
- `config/rabbitmq.conf` — broker config (used only if the email/notification flow needs it).
- `docs/` — design notes (auth flow, monitoring, security). Read before changing those subsystems.

### Template inheritance — modify-in-place mapping

Forked from a microblog gRPC template. The strategy is to **repurpose existing services**, not delete them — every template service has a clear Fiberlane counterpart with similar shape. Keep directory names through the bootstrap to avoid churn; rename only once the new shape is stable.

| Template service | Fiberlane role | What to change |
|---|---|---|
| `api-gateway` | API Gateway | Rewrite routes (`routes.go`). Add clients for new gRPC services. |
| `auth-service` | Buyer auth + supplier magic links | Add `role` enum to JWT claims. Add a magic-link JWT issuer (separate signing key, RFQ-scoped) for supplier RFQ responses — no login required for suppliers. |
| `user-service` | Buyers + suppliers + admins | Add `role` enum (`buyer / supplier / admin`). Suppliers are records without passwords. Trim post-related fields when convenient. |
| `post-service` | **RFQ / Inquiry service** (`post = inquiry/RFQ`) | A "post" becomes an "RFQ": buyer publishes a quote request, suppliers respond. Reuse the existing CRUD + event-publish shape. Rename internal types to `Inquiry` / `Quote` incrementally; the directory and proto package can stay as `post` through the bootstrap to keep imports stable. |
| `notification-service` | Notifications + transactional email | Already a RabbitMQ consumer. Add a Resend sender. Consume `rfq.*` and `order.*` events; route to email (supplier magic-link RFQ invites, buyer quote-received pings, order status changes). |
| `search-service` | Hybrid AI search | Same role (search). **Replace the engine**: drop OpenSearch + Kafka indexer, add Claude tool-use (Haiku 4.5) spec extraction + pgvector ranking against `catalog-service` / Postgres. Kafka topics can be repurposed for embedding re-index jobs later if needed. |
| New: `catalog-service` (was implicit in old `post-service`) | Suppliers + products + embeddings | Either add as a new service or grow inside `post-service`. Default: **new service** so the RFQ flow stays clean. Holds the `vector(1024)` column. |
| New: `order-service` | Order state machine + events | Per GAU-268 + the Order Service Architecture doc. |
| Postgres images | Vector-enabled | **Switch** all per-service Postgres containers to `pgvector/pgvector:pg17` (or `pg16`), and `CREATE EXTENSION IF NOT EXISTS vector` in each init script. |
| OpenSearch, Kafka | Optional / parked | Can be left in compose temporarily; not used by Fiberlane MVP. Disable when not running them locally to save resources. |
| `proto/auth/v1`, `proto/user/v1` | Keep, extend | Add role-related fields, magic-link RPCs. |
| `proto/post/v1` | Repurposed → RFQ contract | Treat its messages as the RFQ/Quote contract. Add fields incrementally; consider renaming the proto package to `inquiry/v1` only after the buyer flow stabilizes (renaming the proto package is invasive — wait until it's worth it). |
| `proto/search/v1` | Keep, rewrite messages | Same package name, different fields (spec chips, match scores, explanation strings). |

New proto domains to add under `proto/` once we split things out: `supplier/v1`, `product/v1` (or a combined `catalog/v1`), `order/v1`. Defer creating these until they have real consumers — don't pre-bake empty contracts.

> Rule of thumb: rename **content**, not **paths**, during the bootstrap. Imports and Docker contexts will fight you if you rename directories too early. Re-evaluate after Day 1.

## Common commands

### Stack (Docker Compose, via `make`)
- `make up-d` / `make down` — start/stop everything detached.
- `make infra-up` — start only infra (redis, per-service postgres, rabbitmq if needed, prometheus, grafana). Use this when running services locally with `go run`.
- `make app-up` / `make app-down` — start/stop only application services.
- `make logs-svc SVC=<name>` — follow logs for one service.
- `make shell SVC=<name>` — shell into a container.
- `make build` / `make build-no-cache` — (re)build images.
- `make clean` — `down -v --remove-orphans` (drops volumes; destroys DB data).

> The Makefile's `infra-up` / `app-up` targets reference all template services. They will keep working as services are repurposed in place — no renames needed in the Makefile during the bootstrap.

### Go services
Each service is a separate module. Always `cd services/<name>` first.
- Build: `go build ./...`
- Test (whole module): `go test ./...`
- Test single package: `go test ./internal/application/services/...`
- Test single test: `go test ./internal/... -run TestName`
- After changing dependencies in any service: `go mod tidy` inside that service directory.
- The `proto` module also has tests: `cd proto && go test ./...`.
- After editing a `.proto` file: regenerate `.pb.go` / `_grpc.pb.go` (committed) and run `go mod tidy` in any consumer service.

### Schema & code generation
- **Migrations** — file-based, golang-migrate. SQL lives in `services/<svc>/migrations/NNN_name.up.sql` + `.down.sql`. Bundled into the service binary via `//go:embed` in `main.go` and applied on boot through `postgres.RunMigrations(db, fs)`. Versions tracked in the `schema_migrations` table inside each service's DB. To add a migration: bump N, write up + down, restart the service (or rebuild the image — the embed.FS is compiled in).
- **sqlc** — `services/<svc>/sqlc.yaml`. Schema is read from `migrations/`, queries from `internal/infrastructure/postgres/queries/*.sql`, output committed at `internal/infrastructure/postgres/sqlcgen/`. Regenerate with `sqlc generate` from the service dir. pgvector columns map to `*pgvector.Vector` via an override in `sqlc.yaml`.
- **protoc** — generated `.pb.go` / `_grpc.pb.go` files are committed alongside each `.proto`. Regenerate from the repo root with:
  ```
  cd proto && protoc -I . \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    <domain>/v1/*.proto
  ```
- **Tooling** (install once): `brew install protobuf` for `protoc`; `go install` for `protoc-gen-go`, `protoc-gen-go-grpc`, `sqlc`. golang-migrate is a Go library — no CLI required for the runtime path.

### Local hybrid run (recommended for backend dev)
1. `make infra-up`
2. `cd services/<name> && go run .` for each service you're iterating on.
3. Compose env vars assume Docker DNS hostnames (`postgres_user`, `redis`, `auth-service`, etc.). Override when running outside the network — e.g. `REDIS_URL=localhost:6379 USER_SERVICE_GRPC_ADDR=localhost:50052 go run .`.

### Frontend
From `frontend/`:
- `bun install`
- `bun run dev` — Next dev server on `:3000`.
- `bun run build` — production build (webpack, not turbopack — see `package.json`).
- `bun run lint` / `bun run typecheck`.

### Monitoring (Prometheus + Grafana)
- Each Go service exposes `GET /metrics` (Prometheus format).
- After `make infra-up`: Prometheus UI at `http://localhost:${PROMETHEUS_PORT:-9090}`, Grafana at `http://localhost:${GRAFANA_PORT:-3001}` (admin creds in `.env.example`).
- Target list + metric names: `docs/monitoring.md`.

## Architecture

### Service topology (target Fiberlane state)

The API Gateway is the only public entry point. It speaks **HTTP/REST to clients** and **gRPC to backend services**.

```
client ─HTTP─► api-gateway :8080
                  │
                  ├─gRPC─► auth-service          :50051   (Redis: tokens, OAuth state, magic-link tokens, auth_code)
                  ├─gRPC─► user-service          :50052   (Postgres: buyers, suppliers, roles)
                  ├─gRPC─► post-service (→ RFQ)  :50053   (Postgres: RFQs/Inquiries, Quotes; publishes rfq.* events)
                  ├─gRPC─► search-service        :50054   (Claude tool-use + pgvector against catalog Postgres)
                  └─(later) order-service        :50055   (Postgres: orders + order_events state machine)
                                  │
                                  └── notification-service consumes rfq.* / order.* → Resend emails
                                          + (new) catalog-service for suppliers/products/embeddings
                                          ── may live inside post-service first, split when justified
```

Naming through Day 1: directories and proto packages keep their template names (`post`, `search`, `notification`). Internal types and route paths use Fiberlane vocabulary (`Inquiry`, `Quote`, `Order`). Rename packages only after the flow is end-to-end.

### Per-service code structure (Clean Architecture, template-standard)

```
services/<svc>/
  main.go
  Dockerfile
  go.mod
  internal/
    application/       # use cases, services, DTOs, application errors
    domain/            # entities, repository interfaces, domain services
    infrastructure/    # postgres (sqlc), redis, rabbitmq, resend, anthropic, voyage — concrete impls
    interfaces/        # grpc/, http/, validators — inbound adapters
    config/            # env loading
  clients/             # outbound gRPC clients to other services
  pkg/{logger,metrics,utils}
```

- **Linear's plan references `internal/{usecase,repository,transport}`** — that is a different naming convention. The template uses `application / domain / infrastructure / interfaces`. **Match the template** so new services look like the existing ones.
- `api-gateway` is flatter: `clients/`, `handlers/`, `middleware/`, `routes/`, `models/`.

### Data layer (Fiberlane-specific)

- **Postgres per service** (no cross-service joins). Image: `pgvector/pgvector:pg16`.
- **Migrations**: `golang-migrate` (`internal/infrastructure/<store>/migrations/*.sql`). Run on service boot, driven by `DB_MIGRATION_PATH` (defaults to `./migrations`). The `scripts/postgres-init-*.sql` files only bootstrap database/role at first container start.
- **Typed queries**: `sqlc` (`sqlc.yaml` per service, generated code committed). Repository interfaces live in `internal/domain/`; sqlc implementations live in `internal/infrastructure/postgres/`.
- **Embeddings**: `vector(1024)` column on `products`, generated by a Go CLI under `cmd/embed/` using Voyage AI `voyage-3` with an OpenAI `text-embedding-3-large` fallback. See GAU-244.
- **Seed data**: real Chinese suppliers + transceiver SKUs (≥80 products) loaded via `cmd/seed/` reading YAML in `seeds/`. See GAU-243.

### Auth model

Building on the template's auth (see `docs/auth-user-management-and-verification.md`). Fiberlane-specific changes:

- **Buyer flow** — unchanged from template: email/password and Google OAuth, JWT access (15 min, in memory on the client) + refresh (14 days, HttpOnly cookie via gateway).
- **Supplier flow** — no login. `auth-service` issues a short-lived **magic-link JWT** (separate issuer key, scoped to a specific RFQ id) embedded in the email link. The supplier RFQ-response page validates this token and accepts a quote without a session.
- **Roles**: `role enum('buyer','supplier','admin')` on the user record. Suppliers have a row but no password.
- **Authorization** still belongs on the receiving service (`actor_id == id` check), not only on the gateway.
- **Auth-code exchange (Google OAuth)**: `GET /api/v1/auth/google` → Google → `GET /api/v1/auth/google/callback` (issues a 5-min `auth_code` in Redis, redirects to client) → `POST /api/v1/auth/exchange` (returns JWT pair). Mobile requires PKCE. State + auth_code use `GETDEL` for one-shot semantics.

> **Carry-over caveat from template**: `DeleteUserTokens` uses `KEYS auth:*:*` — O(N). Don't assume it scales.

### Hybrid AI search (the signature feature — GAU-245)

`search-service` is the heart of Fiberlane. Pipeline:

1. Buyer submits natural-language query through `/api/v1/search` (e.g. *"100G QSFP28 transceiver, 10km, Cisco Nexus compatible"*).
2. **Spec extraction** via Claude Haiku 4.5 using tool-use — returns structured `{ data_rate, form_factor, reach_km, wavelength_nm, vendor_compat[], ... }`.
3. **pgvector ranking** — embed the query (Voyage `voyage-3`), `ORDER BY embedding <=> $1 LIMIT N`, then filter/boost by extracted specs.
4. **Match explanations** — second Claude call per top result, short rationale strings shown in the UI.
5. **Cost target**: under $0.05/query. Cache query → spec-extraction in Redis. Cache embedding by query hash.

The frontend renders editable spec chips that re-run search (GAU-247).

### Order state machine (GAU-268)

State machine + event log lives in `order-service`. Authoritative description is in the Linear doc *Order Service Architecture*. Implementation: domain layer owns the transitions, usecase layer publishes `order_events` rows, gRPC handlers are thin. Order timeline UI (GAU-253) reads `order_events` directly — do **not** denormalize state into the frontend.

### Routing (gateway)

`services/api-gateway/internal/routes/routes.go` is the source of truth for the public API. Target surface for Fiberlane:

- `/api/v1/auth/*` — register/login/google/callback/exchange/refresh (public) + logout/validate (protected). **From template.**
- `/api/v1/search` — protected; hybrid AI search.
- `/api/v1/products/*`, `/api/v1/suppliers/*` — protected reads + admin writes.
- `/api/v1/rfqs/*` — protected. Buyer create + view; supplier responses go through magic-link tokens, not this surface.
- `/api/v1/supplier-rfq/:token` — public, signed-token-gated supplier RFQ response endpoint.
- `/api/v1/orders/*` — protected.
- `/api/v1/admin/*` — protected, admin role required.

### Frontend (Next.js 16 App Router)

The frontend is **not Vite** despite what some Linear issues say (they were drafted against a generic React/Vite template). Use the Next.js conventions that are already wired:

- App Router under `frontend/src/app/`. Auth-gated routes go under `frontend/src/app/(protected)/...`. Public flows (landing, supplier magic-link page, login, register) at the root.
- **BFF** under `frontend/src/app/api/bff/*` — server routes that hold the refresh cookie, attach the access token, and proxy to `api-gateway`. Add a new BFF route for every gateway endpoint the browser calls. Don't call `api-gateway` directly from client components.
- UI primitives in `frontend/src/components/ui/` — `button`, `card`, `input`, `label`, `modal`, `pill`, `badge`. GAU-242 calls for two more: `RouteIndicator` and `CodeId`. Add them here.
- Design tokens (GAU-241): edit `frontend/src/app/globals.css` + Tailwind 4 theme. Industrial palette only:
  - `--ink: #0A0A0A` (near-black), `--paper: #F7F5F0` (off-white linen), `--accent: #D54E20` (international orange), `--cn-vermillion: #C8312D`, `--us-navy: #1A2B4A`. No gradients, no marketing illustrations, no AI-slop imagery.
- State: keep the existing `auth-store` (Zustand). For non-auth state, prefer React Query over global stores.
- Fonts already installed: `@fontsource/jetbrains-mono`, `@fontsource/space-grotesk`. JetBrains Mono is the default for data and identifiers; Space Grotesk for headings.

### Configuration

All runtime config is env-based. Each service has `internal/config/` and reads from `./services/<name>/.env` (mounted via `env_file:` in compose) plus environment overrides. Common knobs: `LOG_LEVEL`, `ENVIRONMENT`, `GRPC_TLS_*`, `GRPC_REFLECTION_ENABLED`, per-service `*_GRPC_ADDR`. mTLS supported but off by default.

Fiberlane-specific env vars:
- `ANTHROPIC_API_KEY` — Claude API for search-service.
- `VOYAGE_API_KEY` / `OPENAI_API_KEY` — embeddings (Voyage primary, OpenAI fallback).
- `RESEND_API_KEY` — transactional email.
- `MAGIC_LINK_JWT_SECRET` / `MAGIC_LINK_TTL` — supplier RFQ response tokens.
- `FRONTEND_URL` — used in email links + OAuth redirect allow-list.

## Conventions to preserve

- **Don't bypass the proto module's replace directive.** New service = separate `go.mod` + `replace github.com/nikitashilov/microblog_grpc/proto => ../../proto`. Dockerfiles build from the **repo root** so they can `COPY proto`; mirror existing `services/<svc>/Dockerfile` layout.
- **Gateway ↔ services is gRPC, not HTTP.** Add cross-service calls via proto contracts, regenerate, wire a client in `services/api-gateway/internal/clients/`.
- **Authorization belongs on the receiving service.** The gateway extracts `userID` from the access token and passes it as `actor_id` in gRPC; the service enforces it (`actor_id == id`, role checks).
- **One Postgres per service.** No cross-service joins; cross-service data flows over gRPC or events.
- **Match the template's Clean Architecture naming** (`application / domain / infrastructure / interfaces`) even when Linear issues use different terms (`usecase / repository / transport`).
- **Real data, real specs.** The demo lives or dies on credible catalog data. Don't fabricate suppliers or transceiver SKUs — source them. (GAU-243.)
- **Design discipline.** No gradients. No marketing illustrations. No emojis in product UI. The route indicator `SHENZHEN ─────► SAN FRANCISCO` is the brand and appears on every relevant screen.
- **Linear is source of truth for scope.** When an issue conflicts with these notes, surface the conflict — don't silently pick one. Conventional cases (Vite paths, `apps/web`, alternate Clean Arch naming) are already reconciled above.

## Day plan (from Linear project)

- **Day 1 — Foundation & Data** (milestone `c288ff2a`): bootstrap from template (GAU-238), auth + roles + magic-link (GAU-239), proto + migrations + sqlc (GAU-240), supplier seed (GAU-243), embeddings CLI (GAU-244), design tokens (GAU-241), UI primitives (GAU-242).
- **Day 2 — AI Search & Buyer Flow** (milestone `cec7f0d4`): hybrid AI search (GAU-245), search home (GAU-246), results with editable chips (GAU-247), product detail (GAU-248), RFQ flow (GAU-249), supplier magic-link page (GAU-250), quote comparison (GAU-251), order-service backend (GAU-268).
- **Day 3 — Polish, Landing & Demo** (milestone `66654e56`): orders dashboard + timeline (GAU-253), landing page (GAU-254), How It Works (GAU-255), supplier profile (GAU-252), supplier application + admin (GAU-256), trust layer (GAU-257), demo prep (GAU-258), outreach list (GAU-259).
