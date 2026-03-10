# Research: Minimal Starter Template

**Branch**: `001-starter-template` | **Date**: 2026-03-10

## Resolved Decisions

### R-001: Session Cookie Security in Development

**Decision**: Use `Secure: false` in development, `Secure: true` in production. Detect via `ENVIRONMENT` env var.

**Rationale**: Browsers block `Secure` cookies over plain HTTP. Development runs on `localhost:3000` / `localhost:8080` without TLS. Deathcookies always sets `Secure: true` but runs behind a TLS-terminating proxy even in staging.

**Alternatives considered**:
- Always set `Secure: true` — breaks local development without mkcert/TLS setup
- Use `SameSite=None` — less secure, requires Secure flag anyway

### R-002: Frontend-to-Backend Communication Pattern

**Decision**: Next.js catch-all API route (`/api/v1/[...path]/route.ts`) proxies requests to Go backend. Frontend never calls Go directly from the browser.

**Rationale**: Same pattern as deathcookies. The proxy lets the Next.js server add server-side auth headers and keeps the Go backend's URL internal. Cookies are forwarded transparently.

**Alternatives considered**:
- Browser calls Go directly — requires CORS configuration between ports, exposes backend URL to client, loses server-side auth header injection
- Next.js Server Actions — would work for mutations but doesn't fit the REST API pattern and would diverge from deathcookies conventions

### R-003: Hot Reload Strategy for Development

**Decision**: `make dev` runs Go server via `go run` (simplest). Air (hot reload) is optional and can be added later.

**Rationale**: For a minimal starter, `go run` is sufficient. Air adds a dependency and config file. The Makefile target can be swapped to Air later without changing the developer workflow (still `make dev`).

**Alternatives considered**:
- Air from the start — adds `.air.toml` config, binary dependency. Premature for a starter template.
- CompileDaemon — less popular, same trade-offs as Air

### R-004: Database Migration Runner

**Decision**: Use `golang-migrate` CLI via Makefile targets. Migrations are up-only (no down files), consistent with CLAUDE.md's "migrations are irreversible" philosophy.

**Rationale**: Deathcookies uses golang-migrate with 61 migrations. Up-only aligns with the expand-and-contract pattern prescribed by the project constitution.

**Alternatives considered**:
- goose — similar capabilities, but diverges from deathcookies tooling
- Atlas — more powerful (declarative), but heavier and unfamiliar to the project

### R-005: sqlc Configuration

**Decision**: sqlc v2 with PostgreSQL engine. Queries in `internal/repository/queries/`, output to `internal/repository/sqlcgen/`. Standard type mappings: `timestamptz → pgtype.Timestamptz`, `serial → int32`, `text → string`, `boolean → bool`.

**Rationale**: Matches deathcookies sqlc.yaml configuration exactly. Type-safe generated code eliminates SQL injection and mapping bugs.

**Alternatives considered**:
- GORM or sqlx — ORMs/query builders add abstraction; sqlc generates plain Go from plain SQL, which is simpler and matches project conventions
- pgx raw queries — loses type safety and requires manual scanning

### R-006: Password Hashing Cost

**Decision**: bcrypt cost 12, matching deathcookies.

**Rationale**: Cost 12 provides ~300ms computation time, making brute-force attacks expensive while remaining acceptable for login latency. Industry standard for web applications.

**Alternatives considered**:
- bcrypt cost 10 (default) — faster but less secure
- argon2id — newer, more resistant to GPU attacks, but adds dependency and diverges from deathcookies pattern

### R-008: Rate Limiter Implementation

**Decision**: In-memory rate limiter (Go sync.Mutex + map). Single-instance only.

**Rationale**: The starter template runs as a single process. An in-memory rate limiter is zero-dependency and sufficient. The limitation (resets on restart, no cross-instance sync) is documented in the type contract and acceptable for a single-owner app.

**Alternatives considered**:
- Redis-backed limiter — adds Redis dependency. Correct for production multi-instance, but premature for a starter template.
- Database-backed limiter — adds a table and queries for a cross-cutting concern. Slower than in-memory.
- No rate limiting — leaves login endpoint open to brute-force. Unacceptable even for a starter.

### R-009: CORS Configuration Default

**Decision**: Default CORS origin is `http://localhost:3000` (the Next.js dev server), not `*`.

**Rationale**: The application uses cookie-based authentication with `Access-Control-Allow-Credentials: true`. The Fetch specification forbids combining `Allow-Credentials: true` with `Allow-Origin: *`. Using `*` would cause browsers to reject credentialed responses. The default must be a specific origin.

**Alternatives considered**:
- Wildcard `*` — incompatible with `Allow-Credentials: true`. Would break cookie-based auth.
- No CORS (same-origin only) — would work if the proxy handles everything, but the backend may be called directly during development and testing.

### R-007: Frontend Package Manager

**Decision**: npm (matches deathcookies `package-lock.json`).

**Rationale**: Consistency with reference project. No compelling reason to switch for a starter template.

**Alternatives considered**:
- pnpm — faster, stricter, but diverges from deathcookies
- bun — fastest, but less mature ecosystem
