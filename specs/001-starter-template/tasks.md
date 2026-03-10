# Tasks: Minimal Starter Template

**Input**: Design documents from `/specs/001-starter-template/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module, Next.js app, build tooling

- [ ] T001 Initialize Go module with `go.mod` (module path: github.com/yomauro or appropriate) at project root
- [ ] T002 Create sqlc configuration in `sqlc.yaml` per research R-005 (PostgreSQL engine, queries in `internal/repository/queries/`, output to `internal/repository/sqlcgen/`)
- [ ] T003 Create `.env.example` at project root with all documented environment variables per plan Configuration section
- [ ] T004 Create `Makefile` at project root with targets: dev, build, migrate, migrate-create, sqlc, web-dev, web-build, setup per plan Makefile Targets section
- [ ] T005 [P] Initialize Next.js 14 project in `web/` with App Router, TypeScript, Tailwind CSS 3.4, and path alias `@/*` mapping to `./src/*`
- [ ] T006 [P] Create `web/tailwind.config.ts` with custom spacing (`sidebar: "16rem"`), CSS variable colors, and typography plugin
- [ ] T007 [P] Create `web/src/app/globals.css` with Tailwind directives and CSS variables (`--background: #f9fafb`, `--foreground: #111827`)
- [ ] T008 [P] Create `web/src/app/layout.tsx` root layout with Inter font (Google Fonts), viewport meta, and globals.css import

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema, sqlc codegen, Go packages (boundary files first), config loading — MUST complete before any user story

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T009 Create migration `migrations/000001_users.up.sql` with users table per data-model.md (id serial PK, email text UNIQUE NOT NULL, password_hash, display_name, role CHECK IN ('owner'), active, created_at, updated_at)
- [ ] T010 Create migration `migrations/000002_sessions.up.sql` with sessions table per data-model.md (id text PK, user_id FK CASCADE, created_at, expires_at) plus indexes on user_id and expires_at
- [ ] T011 Create sqlc queries in `internal/repository/queries/auth.sql` per contracts/types.md: GetUserByEmail, CountActiveOwners, CreateUser, CreateSession, GetSession (join with active check), ExtendSession, DeleteSession, DeleteUserSessions, DeleteExpiredSessions
- [ ] T012 Run `make sqlc` to generate Go code in `internal/repository/sqlcgen/`
- [ ] T013 [P] Create boundary file `internal/config/config.go` with Settings struct and LoadSettings() per contracts/types.md (DatabaseURL, ListenAddr, AuthSecret, CORSOrigins default "http://localhost:3000", AdminEmail, AdminPassword, Environment default "development", IsDev() helper, non-dev validation)
- [ ] T014 [P] Create boundary file `internal/auth/auth.go` with exported functions per contracts/types.md: HashPassword (bcrypt cost 12), CheckPassword, DummyHash (sync.Once), GenerateSessionToken (32 bytes hex), RoleLevel (owner=3), and RateLimiter type with NewRateLimiter and Allow
- [ ] T015 Create `internal/auth/ratelimit.go` with RateLimiter implementation (sync.Mutex + map, time-windowed, per contracts/types.md limitation note)
- [ ] T016 Create `internal/auth/bootstrap.go` with BootstrapOwner implementation: check CountActiveOwners, skip if >0 or email/password empty, hash password, CreateUser with role=owner
- [ ] T017 Create boundary file `internal/handler/handler.go` with Handler struct, AuthUser struct, GetAuthUser() context helper, contextKey type, and NewRouter() function signature per contracts/types.md
- [ ] T018 Create `internal/handler/middleware.go` with SessionAuthMiddleware (read ym_session cookie, GetSession query, check active, extend session 7 days, inject AuthUser into context, return 401/503 on failure), plus RequestID, Recovery, Logging (slog), and CORS (Allow-Credentials: true, specific origin) middleware
- [ ] T019 Create `internal/handler/auth.go` with Login handler (rate limit, parse JSON, GetUserByEmail, DummyHash timing prevention, bcrypt check, CreateSession, set ym_session cookie with HttpOnly/Secure/SameSite=Lax/MaxAge=604800, return user JSON), Logout handler (DeleteSession, clear cookie), Me handler (return AuthUser from context), and Health handler (pool.Ping, return 200/503) per contracts/api.md
- [ ] T020 Wire NewRouter() in `internal/handler/handler.go`: apply middleware stack (RequestID → Recovery → Logging → CORS → MaxBytesReader), mount public routes (POST /api/v1/auth/login, GET /api/v1/health), mount protected routes with SessionAuthMiddleware (GET /api/v1/auth/me, POST /api/v1/auth/logout)
- [ ] T021 Create `cmd/api/main.go` entrypoint: LoadSettings, pgxpool.New, run migrations check, DeleteExpiredSessions on startup, BootstrapOwner, NewRouter, http.Server with ListenAddr, graceful shutdown (SIGINT/SIGTERM, 10s timeout, pool.Close), structured slog logger

**Checkpoint**: Backend fully functional — all API endpoints work, owner bootstrapped, sessions managed. Can verify with curl per quickstart QS-016 through QS-020.

---

## Phase 3: User Story 1 - First-Time Setup and Login (Priority: P1) MVP

**Goal**: End-to-end flow: `make setup` → `make dev` → open browser → login → see "Yo Mauro" dashboard

**Independent Test**: Run `make setup && make dev`, open browser, login with admin credentials, verify dashboard shows "Yo Mauro" in large text

### Implementation for User Story 1

- [ ] T022 [P] [US1] Create `web/src/lib/api.ts` with typed fetch wrapper: ApiError class, NetworkError class, generic request/get/post functions, api object with login(), logout(), me() methods per contracts/types.md
- [ ] T023 [P] [US1] Create `web/src/lib/auth.ts` with isPublicPath() helper and PUBLIC_PATHS set (/login, /api/*, /_next/*, /healthz)
- [ ] T024 [US1] Create `web/src/middleware.ts` with session validation: check ym_session cookie, validate against /api/v1/auth/me via INTERNAL_API_URL, cache user in ym_user cookie (5-min TTL, no email per security review), set x-user-id and x-user-role headers, redirect to /login on failure, clear both cookies on failure
- [ ] T025 [US1] Create `web/src/app/api/v1/[...path]/route.ts` catch-all API proxy: forward requests to INTERNAL_API_URL, forward Cookie header, return Set-Cookie headers from backend
- [ ] T026 [P] [US1] Create `web/src/app/(auth)/layout.tsx` auth layout: centered flex container, light gray bg, "Yo Mauro" branding text, max-width 448px
- [ ] T027 [US1] Create `web/src/components/auth/LoginForm.tsx` client component: email + password fields with HTML required attributes, submit handler POST to /api/v1/auth/login via api client, error state display ("invalid email or password" for ApiError, "Unable to connect to server" for NetworkError), router.push("/") + router.refresh() on success
- [ ] T028 [US1] Create `web/src/app/(auth)/login/page.tsx` rendering LoginForm component
- [ ] T029 [P] [US1] Create `web/src/app/(dashboard)/page.tsx` dashboard page with large centered "Yo Mauro" text (prominent heading, centered both vertically and horizontally in main content area)
- [ ] T030 [US1] Create `web/src/components/layout/Sidebar.tsx` client component: desktop fixed left sidebar (16rem, hidden below md), Dashboard link with active highlight (bg-gray-900 text-white), Logout button at bottom with border-t separator, logout calls api.logout() then router.push("/login") + router.refresh()
- [ ] T031 [US1] Create `web/src/app/(dashboard)/layout.tsx` dashboard layout: render Sidebar + main content area with responsive padding
- [ ] T032 [US1] Create `web/.env.local` with NEXT_PUBLIC_API_URL=http://localhost:8080 and INTERNAL_API_URL=http://localhost:8080

**Checkpoint**: Full US1 flow works — make setup, make dev, make web-dev, login, see "Yo Mauro" dashboard. Verify with quickstart QS-001 through QS-006.

---

## Phase 4: User Story 2 - Session Persistence and Logout (Priority: P2)

**Goal**: Sessions persist across browser restarts (7-day sliding window), logout works, expired sessions redirect to login

**Independent Test**: Login, close tab, reopen — dashboard loads without login. Click logout — redirected to login. Delete session from DB — next navigation redirects to login.

### Implementation for User Story 2

- [ ] T033 [US2] Verify session sliding window in `internal/handler/middleware.go` — ExtendSession call extends expires_at by 7 days on each authenticated request (already implemented in T018, verify behavior)
- [ ] T034 [US2] Verify logout flow end-to-end: Sidebar logout button → api.logout() → POST /api/v1/auth/logout → DeleteSession + clear cookie → redirect to /login (already wired in T030, verify behavior)
- [ ] T035 [US2] Verify stale cookie handling in `web/src/middleware.ts`: when ym_session cookie exists but session is deleted from DB, middleware gets 401 from /auth/me, clears both ym_session and ym_user cookies, redirects to /login (already implemented in T024, verify behavior)
- [ ] T036 [US2] Verify ym_user cache cookie expiry: when cache expires (5-min TTL), middleware re-validates against backend, session sliding window keeps session alive (verify T024 cache logic)

**Checkpoint**: Session persistence and logout verified. Quickstart QS-009 through QS-013 pass.

---

## Phase 5: User Story 3 - Mobile Navigation (Priority: P3)

**Goal**: Bottom tab bar on mobile viewports (<768px) instead of sidebar

**Independent Test**: Open app on mobile viewport, verify bottom tab bar with Dashboard and Logout, tap logout — redirected to login.

### Implementation for User Story 3

- [ ] T037 [US3] Add mobile bottom tab bar to `web/src/components/layout/Sidebar.tsx`: fixed bottom nav (56px min height), Dashboard + Logout as stacked icon+text items, safe area inset for notched devices (env(safe-area-inset-bottom)), visible only below md breakpoint
- [ ] T038 [US3] Verify responsive behavior: sidebar hidden below md, bottom tab bar hidden at md+, dashboard content area adjusts padding for bottom bar on mobile

**Checkpoint**: Mobile navigation works. Quickstart QS-014 and QS-015 pass.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Security hardening, logging, cleanup

- [ ] T039 Add security event logging in `internal/handler/auth.go`: slog.Warn for failed login attempts (IP + email, no password), rate limit triggers (IP), session validation failures (session ID prefix only) per plan Cross-Cutting Concerns
- [ ] T040 Add request body size limit middleware in `internal/handler/handler.go` NewRouter: http.MaxBytesReader 1MB for all routes that read request bodies
- [ ] T041 Verify graceful shutdown in `cmd/api/main.go`: SIGINT/SIGTERM handling, 10s shutdown timeout, pool.Close(), shutdown log message
- [ ] T042 Update CLAUDE.md Stack Layers table with yomauro-specific paths (remove example references to analytics/classify, add actual paths)
- [ ] T043 Run full quickstart validation (QS-001 through QS-020) per specs/001-starter-template/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase — delivers MVP
- **User Story 2 (Phase 4)**: Depends on US1 (session/logout already implemented, this phase verifies behavior)
- **User Story 3 (Phase 5)**: Depends on US1 (extends Sidebar component)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational (Phase 2) — core end-to-end flow
- **User Story 2 (P2)**: Depends on US1 — verifies session behavior already built in US1
- **User Story 3 (P3)**: Depends on US1 — extends the Sidebar component built in US1

### Within Each Phase

- Boundary files before implementation files
- Migrations before sqlc codegen
- Backend packages before frontend (frontend calls backend API)
- [P] tasks within a phase can run in parallel

### Parallel Opportunities

- T005, T006, T007, T008 (frontend setup) can run in parallel with T001-T004 (backend setup)
- T013, T014 (boundary files) can run in parallel after T012
- T022, T023, T026, T029 can run in parallel within US1
- US2 and US3 could theoretically run in parallel after US1 (different concerns)

---

## Parallel Example: Phase 2 Foundational

```
# After T012 (sqlc generate), launch boundary files in parallel:
T013: Create internal/config/config.go (boundary)
T014: Create internal/auth/auth.go (boundary)

# After boundaries, implementation files in parallel:
T015: Create internal/auth/ratelimit.go
T016: Create internal/auth/bootstrap.go
```

## Parallel Example: User Story 1

```
# After Phase 2 complete, launch parallel frontend tasks:
T022: Create web/src/lib/api.ts
T023: Create web/src/lib/auth.ts
T026: Create web/src/app/(auth)/layout.tsx
T029: Create web/src/app/(dashboard)/page.tsx
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T008)
2. Complete Phase 2: Foundational (T009-T021)
3. Complete Phase 3: User Story 1 (T022-T032)
4. **STOP and VALIDATE**: Login → "Yo Mauro" dashboard works end-to-end
5. This IS the MVP — a working authenticated app

### Incremental Delivery

1. Setup + Foundational → Backend API works (curl verification)
2. Add User Story 1 → Full login-to-dashboard flow (MVP!)
3. Add User Story 2 → Session persistence verified
4. Add User Story 3 → Mobile navigation complete
5. Polish → Security hardening, logging, full quickstart validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US2 is primarily a verification phase — most behavior is built in US1's implementation
- US3 extends the Sidebar component from US1 — not a separate component
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
