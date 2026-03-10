# API Contracts: Minimal Starter Template

**Branch**: `001-starter-template` | **Date**: 2026-03-10

## Base URL

All endpoints under `/api/v1/`. Frontend proxies requests through Next.js catch-all route.

## Cross-Cutting API Constraints

**Request body size limit**: All endpoints enforce a 1 MB maximum request body via `http.MaxBytesReader`. Requests exceeding this limit receive HTTP 413.

**CORS**: The Go backend sets `Access-Control-Allow-Origin` to the configured `CORS_ORIGINS` value (default `http://localhost:3000` in development). Wildcard `*` is prohibited because `Access-Control-Allow-Credentials: true` is required for cookie-based auth. Production deployments must set `CORS_ORIGINS` to the actual frontend origin.

**CSRF**: The `SameSite=Lax` attribute on the `ym_session` cookie prevents cross-origin POST requests from external sites. Combined with the JSON `Content-Type` requirement (which triggers a CORS preflight for cross-origin requests), this provides adequate CSRF protection without a separate token mechanism.

## Authentication Endpoints

### POST /api/v1/auth/login

**Auth**: Public (rate-limited: 5 attempts/IP/minute)

**Request**:
```json
{
  "email": "admin@example.com",
  "password": "changeme"
}
```

**Response 200** (success):
```json
{
  "user": {
    "id": 1,
    "email": "admin@example.com",
    "display_name": "Owner",
    "role": "owner"
  }
}
```

**Side effect**: Sets `ym_session` cookie (HttpOnly, Secure in prod, SameSite=Lax, MaxAge=604800).

**Response 400** (missing fields):
```json
{
  "error": "email and password are required"
}
```

**Response 401** (invalid credentials):
```json
{
  "error": "invalid email or password"
}
```

**Response 429** (rate limited):
```json
{
  "error": "too many login attempts, try again later"
}
```

---

### GET /api/v1/auth/me

**Auth**: Requires valid `ym_session` cookie (SessionAuthMiddleware)

**Response 200**:
```json
{
  "id": 1,
  "email": "admin@example.com",
  "display_name": "Owner",
  "role": "owner"
}
```

**Side effect**: Extends session expiry by 7 days (sliding window).

**Response 401**:
```json
{
  "error": "not authenticated"
}
```

---

### POST /api/v1/auth/logout

**Auth**: Requires valid `ym_session` cookie (SessionAuthMiddleware)

**Response 200**:
```json
{
  "message": "logged out"
}
```

**Side effect**: Deletes session row from database. Clears `ym_session` cookie (MaxAge=-1).

---

## Operational Endpoints

### GET /api/v1/health

**Auth**: Public (no authentication required)

**Response 200** (healthy):
```json
{
  "status": "ok"
}
```

**Response 503** (database unreachable):
```json
{
  "status": "unhealthy",
  "error": "database ping failed"
}
```

## Error Responses for Database Unavailability

Any authenticated endpoint may return HTTP 503 when the database is unreachable:
```json
{
  "error": "service temporarily unavailable"
}
```

The login endpoint returns 503 (not 500) for database errors to distinguish infrastructure failures from authentication failures.

## Cookie Contracts

### ym_session (Backend → Browser)

| Attribute | Value |
|-----------|-------|
| Name | `ym_session` |
| Value | 64-character hex string (crypto-random) |
| Path | `/` |
| MaxAge | 604800 (7 days) |
| HttpOnly | true |
| Secure | true in production, false in development |
| SameSite | Lax |

### ym_user (Next.js Middleware → Browser)

| Attribute | Value |
|-----------|-------|
| Name | `ym_user` |
| Value | JSON-encoded `{id, role, display_name}` (no PII -- email excluded) |
| Path | `/` |
| MaxAge | 300 (5 minutes) |
| HttpOnly | false (read by Next.js middleware on subsequent server-side requests) |
| Secure | true in production, false in development |
| SameSite | Lax |

**Security note**: Because `ym_user` is not HttpOnly, it is readable by client-side JavaScript. It must never contain sensitive data (email, tokens). Only routing-decision fields (id, role, display_name) are included. The `display_name` value must be URL-encoded when written to the cookie to handle special characters safely.

## Middleware Request Headers

Set by Next.js middleware after session validation, available to server components:

| Header | Value | Example |
|--------|-------|---------|
| `x-user-id` | User's numeric ID | `1` |
| `x-user-role` | User's role string | `owner` |
