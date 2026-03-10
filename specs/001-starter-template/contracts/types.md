# Type Contracts: Minimal Starter Template

**Branch**: `001-starter-template` | **Date**: 2026-03-10

## Go Types

### internal/auth/auth.go (Boundary)

```go
// Package auth provides password hashing, session token generation,
// rate limiting, and owner bootstrapping for the authentication system.
package auth

// RateLimiter tracks login attempts per IP address with time-based expiry.
// LIMITATION: In-memory only. Resets on server restart and does not synchronize
// across multiple instances. Acceptable for a single-instance starter template.
// Replace with Redis-backed limiter if horizontal scaling is needed.
type RateLimiter struct { /* unexported fields */ }

// NewRateLimiter creates a rate limiter with the given max attempts and time window.
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter

// Allow returns true if the key (IP address) has not exceeded the rate limit.
func (rl *RateLimiter) Allow(key string) bool

// HashPassword hashes a plaintext password with bcrypt cost 12.
func HashPassword(password string) (string, error)

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(password, hash string) error

// DummyHash returns a pre-computed bcrypt hash for timing-attack prevention.
// Use when the user is not found so response time is constant.
func DummyHash() string

// GenerateSessionToken returns a crypto-random 32-byte hex-encoded string (64 chars).
func GenerateSessionToken() (string, error)

// RoleLevel returns the numeric hierarchy level for a role string.
// owner=3. Returns 0 for unknown roles.
func RoleLevel(role string) int

// BootstrapOwner creates the first owner account if no active owners exist.
// No-op if an active owner already exists or if email/password are empty.
func BootstrapOwner(ctx context.Context, q *sqlcgen.Queries, email, password string, logger *slog.Logger) error
```

### internal/config/config.go (Boundary)

```go
// Package config provides environment-based application settings.
package config

// Settings holds all configuration values loaded from environment variables.
type Settings struct {
    DatabaseURL   string // PostgreSQL connection string
    ListenAddr    string // HTTP server listen address (default ":8080")
    AuthSecret    string // Shared secret for internal auth (not used in minimal version, reserved)
    CORSOrigins   string // Allowed CORS origins (default "http://localhost:3000" in dev)
    AdminEmail    string // Bootstrap owner email
    AdminPassword string // Bootstrap owner password
    Environment   string // "development", "staging", "production" (default "development")
}
// NOTE on CORS: The default is "http://localhost:3000" (the Next.js dev server), NOT "*".
// A wildcard origin ("*") cannot be used with Access-Control-Allow-Credentials: true,
// which is required for cookie-based authentication. Production must set this explicitly.

// LoadSettings reads configuration from environment variables.
// In non-development environments, validates that critical settings are not defaults.
func LoadSettings() (*Settings, error)

// IsDev returns true if Environment is "development".
func (s *Settings) IsDev() bool
```

### internal/handler/handler.go (Boundary)

```go
// Package handler provides HTTP request handlers and middleware
// for the yomauro API.
package handler

// AuthUser is the user data stored in request context by SessionAuthMiddleware.
type AuthUser struct {
    ID          int32  `json:"id"`
    Email       string `json:"email"`
    DisplayName string `json:"display_name"`
    Role        string `json:"role"`
}

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct { /* unexported fields: pool, queries, settings, logger, loginLimiter */ }

// NewRouter creates a chi router with all middleware and routes configured.
func NewRouter(pool *pgxpool.Pool, queries *sqlcgen.Queries, settings *config.Settings, logger *slog.Logger) chi.Router

// GetAuthUser extracts the authenticated user from request context.
// Returns nil if no user is present (unauthenticated request).
func GetAuthUser(ctx context.Context) *AuthUser
```

## sqlc Query Signatures

### internal/repository/queries/auth.sql

```sql
-- name: GetUserByEmail :one
-- Returns: id, email, password_hash, display_name, role, active
-- Note: only selects fields needed for authentication; avoids leaking full row into handler scope.
SELECT id, email, password_hash, display_name, role, active FROM users WHERE email = @email;

-- name: CountActiveOwners :one
-- Returns: count (int64)
SELECT COUNT(*) FROM users WHERE role = 'owner' AND active = true;

-- name: CreateUser :one
-- Returns: id, email, password_hash, display_name, role, active, created_at, updated_at
INSERT INTO users (email, password_hash, display_name, role)
VALUES (@email, @password_hash, @display_name, @role)
RETURNING *;

-- name: CreateSession :one
-- Returns: id, user_id, created_at, expires_at
INSERT INTO sessions (id, user_id, expires_at)
VALUES (@id, @user_id, @expires_at)
RETURNING *;

-- name: GetSession :one
-- Returns: s.id, s.user_id (use as AuthUser.ID), s.created_at, s.expires_at,
--          u.email, u.display_name, u.role
-- Note: checks BOTH session expiry AND user active status. An inactive user
-- with a valid session token must be rejected by middleware.
-- Note: u.id is NOT selected; use s.user_id to populate AuthUser.ID.
SELECT s.id, s.user_id, s.created_at, s.expires_at,
       u.email, u.display_name, u.role
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.id = @id AND s.expires_at > NOW() AND u.active = true;

-- name: ExtendSession :exec
UPDATE sessions SET expires_at = @expires_at WHERE id = @id;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = @id;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = @user_id;

-- name: DeleteExpiredSessions :exec
-- Cleanup: removes expired session rows to prevent unbounded table growth.
-- Called on server startup and optionally on a periodic schedule.
DELETE FROM sessions WHERE expires_at < NOW();
```

## TypeScript Types

### web/src/lib/api.ts

```typescript
// API response types

export interface AuthUser {
  id: number;
  email: string;
  display_name: string;
  role: string;
}

export interface LoginResponse {
  user: AuthUser;
}

export interface LogoutResponse {
  message: string;
}

export interface ErrorResponse {
  error: string;
}

// API error class for non-2xx responses
export class ApiError extends Error {
  constructor(
    public status: number,
    public body: ErrorResponse,
  ) {
    super(body.error);
    this.name = "ApiError";
  }
}

// Network error class for failed fetch (backend unreachable).
// Wraps TypeError thrown by fetch when the server cannot be reached.
export class NetworkError extends Error {
  constructor(cause?: Error) {
    super("Unable to connect to server");
    this.name = "NetworkError";
    this.cause = cause;
  }
}

// API client methods.
// Throws ApiError for non-2xx HTTP responses.
// Throws NetworkError when the backend is unreachable (fetch TypeError).
export const api = {
  login(email: string, password: string): Promise<LoginResponse>;
  logout(): Promise<LogoutResponse>;
  me(): Promise<AuthUser>;
};
```

### web/src/middleware.ts

```typescript
// Cached user data structure (stored in ym_user cookie).
// NOTE: This cookie is NOT HttpOnly (readable by Next.js middleware on the server side
// via request headers). It must NOT contain PII like email. Only the minimum fields
// needed for middleware routing decisions are included.
interface CachedUser {
  id: number;
  role: string;
  display_name: string;
}
```
