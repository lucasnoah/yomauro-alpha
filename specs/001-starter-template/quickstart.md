# Quickstart / Test Scenarios: Minimal Starter Template

**Branch**: `001-starter-template` | **Date**: 2026-03-10

## Prerequisites

Before starting, ensure you have installed:
- Go 1.25+
- Node.js 20+ and npm
- PostgreSQL (running locally or accessible via connection string)
- Make
- golang-migrate CLI (`brew install golang-migrate` on macOS)

## Setup Verification

### QS-001: Fresh Setup

1. Clone the repo: `git clone <repo-url> && cd yomauro`
2. Copy the example env file: `cp .env.example config/.env`
3. Edit `config/.env` and set:
   - `DATABASE_URL` to a fresh PostgreSQL database (e.g., `postgres://localhost:5432/yomauro?sslmode=disable`)
   - `ADMIN_EMAIL` to the desired owner email (e.g., `admin@example.com`)
   - `ADMIN_PASSWORD` to the desired owner password (e.g., `changeme`)
4. Run `make setup` (this installs Go and Node dependencies, runs migrations, generates sqlc code, and bootstraps the owner account)
5. **Verify**: No errors. Database has `users` and `sessions` tables. One owner row exists in `users` matching ADMIN_EMAIL.

### QS-002: Start Both Servers

The backend and frontend must both be running for the application to work.

1. In one terminal: `make dev` (starts Go backend on :8080)
2. In another terminal: `make web-dev` (starts Next.js frontend on :3000)
3. **Verify**: Backend logs "listening on :8080". Frontend compiles without errors. Both terminals remain running.

### QS-002a: Health Check

1. With the backend running, execute:
   ```bash
   curl -s http://localhost:8080/api/v1/health
   ```
2. **Verify**: HTTP 200, response body `{"status":"ok"}`. If the database is down, returns HTTP 503.

## Authentication Scenarios

### QS-003: Login Success

1. Navigate to `http://localhost:3000`
2. **Verify**: Redirected to `/login`
3. Enter the admin email and password configured in `config/.env`, click submit
4. **Verify**: Redirected to dashboard. Page shows "Yo Mauro" in large text. Sidebar visible on desktop.

### QS-004: Login Failure -- Invalid Credentials

1. Navigate to `/login`
2. Enter valid email but wrong password, click submit
3. **Verify**: Error message displayed ("invalid email or password"). Still on login page.

### QS-005: Login Failure -- Non-Existent User

1. Navigate to `/login`
2. Enter non-existent email and any password, click submit
3. **Verify**: Same error message as QS-004 ("invalid email or password"). Response time similar to QS-004 (timing-attack prevention via DummyHash).

### QS-006: Login Failure -- Empty Fields

1. Navigate to `/login`
2. Click submit without entering email or password
3. **Verify**: Form validation prevents submission (client-side HTML required attributes). If bypassed, server returns 400 "email and password are required".

### QS-007: Rate Limiting

1. Submit 5 incorrect login attempts within 1 minute from the same IP (the rate limit window is 1 minute, 5 attempts max):
   ```bash
   for i in $(seq 1 6); do
     curl -s -o /dev/null -w "Attempt $i: HTTP %{http_code}\n" \
       -X POST http://localhost:8080/api/v1/auth/login \
       -H "Content-Type: application/json" \
       -d '{"email":"admin@example.com","password":"wrong"}'
   done
   ```
2. **Verify**: Attempts 1-5 return HTTP 401. Attempt 6 returns HTTP 429 with body `{"error":"too many login attempts, try again later"}`.

### QS-008: Backend Unreachable

1. Stop the Go backend (Ctrl+C in the backend terminal)
2. Navigate to `/login` in the browser, enter credentials, click submit
3. **Verify**: The login form displays "Unable to connect to server" (not a generic error).
4. Restart the backend with `make dev`.

## Session Scenarios

### QS-009: Session Persistence

1. Complete QS-003 (successful login)
2. Close browser tab
3. Open new tab, navigate to `http://localhost:3000`
4. **Verify**: Dashboard loads directly without login prompt.

### QS-010: Logout

1. From the dashboard (logged in)
2. Click "Logout" in sidebar (desktop) or bottom tab bar (mobile)
3. **Verify**: Redirected to `/login`. Navigating to `/` redirects back to `/login`.

### QS-011: Expired Session

1. Login successfully
2. Manually delete the session row from the database to simulate expiry:
   ```bash
   psql "$DATABASE_URL" -c "DELETE FROM sessions;"
   ```
3. Refresh the dashboard page
4. **Verify**: Redirected to `/login`.

### QS-012: Stale Cookie

1. Login successfully, note the `ym_session` cookie value
2. Delete the session row from the database:
   ```bash
   psql "$DATABASE_URL" -c "DELETE FROM sessions;"
   ```
3. Navigate to a protected page
4. **Verify**: Redirected to `/login`. `ym_session` and `ym_user` cookies are cleared.

### QS-013: Inactive User Rejection

1. Login successfully
2. Deactivate the user in the database:
   ```bash
   psql "$DATABASE_URL" -c "UPDATE users SET active = false WHERE email = 'admin@example.com';"
   ```
3. Refresh the dashboard page (or wait for the `ym_user` cache cookie to expire, up to 5 minutes)
4. **Verify**: Redirected to `/login`. The session is no longer valid because `GetSession` checks `u.active = true`.
5. Re-activate for subsequent tests:
   ```bash
   psql "$DATABASE_URL" -c "UPDATE users SET active = true WHERE email = 'admin@example.com';"
   ```

## Navigation Scenarios

### QS-014: Desktop Sidebar

1. Login on a desktop viewport (width >= 768px)
2. **Verify**: Left sidebar visible with "Dashboard" link (active/highlighted) and "Logout" button at bottom.

### QS-015: Mobile Bottom Tab Bar

1. Login on a mobile viewport (width < 768px)
2. **Verify**: No left sidebar. Bottom tab bar visible with "Dashboard" and "Logout" options.

### QS-016: Dashboard Content

1. Login successfully
2. **Verify**: Dashboard page displays "Yo Mauro" in large, prominent text. Text is centered and clearly the focal point of the page.

## API Direct Verification

### QS-017: POST /api/v1/auth/login (Direct)

```bash
curl -v -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"changeme"}' \
  -c cookies.txt
```

**Verify**: HTTP 200, JSON body with user object, `ym_session` cookie set in cookies.txt.

### QS-018: GET /api/v1/auth/me (Direct)

```bash
curl -v http://localhost:8080/api/v1/auth/me -b cookies.txt
```

**Verify**: HTTP 200, JSON body with `{id, email, display_name, role}`.

### QS-019: POST /api/v1/auth/logout (Direct)

```bash
curl -v -X POST http://localhost:8080/api/v1/auth/logout -b cookies.txt -c cookies.txt
```

**Verify**: HTTP 200, `{"message":"logged out"}`. Subsequent `/auth/me` call returns 401.

### QS-020: GET /api/v1/health (Direct)

```bash
curl -v http://localhost:8080/api/v1/health
```

**Verify**: HTTP 200, `{"status":"ok"}`. No authentication required.
