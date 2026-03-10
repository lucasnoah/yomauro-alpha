# Data Model: Minimal Starter Template

**Branch**: `001-starter-template` | **Date**: 2026-03-10

## Entities

### User

Represents an authenticated account in the system.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | serial (auto-increment integer) | PRIMARY KEY | Internal identifier |
| email | text | NOT NULL, UNIQUE | Login credential, unique across system |
| password_hash | text | NOT NULL | bcrypt hash of user's password (never stored plaintext) |
| display_name | text | NOT NULL | Human-readable name shown in UI |
| role | text | NOT NULL, DEFAULT 'owner', CHECK IN ('owner') | Authorization role (single value for now, expandable) |
| active | boolean | NOT NULL, DEFAULT true | Soft-delete flag; inactive users cannot log in |
| created_at | timestamptz | NOT NULL, DEFAULT NOW() | Account creation timestamp |
| updated_at | timestamptz | NOT NULL, DEFAULT NOW() | Last modification timestamp (set on INSERT only; no auto-update trigger in this version since no UPDATE queries exist) |

**Indexes**: Primary key on `id`, unique constraint on `email`.

**Validation Rules**:
- `email` must be non-empty and unique
- `password_hash` is always a bcrypt hash (never plaintext)
- `role` is constrained to `'owner'` via CHECK (expandable by altering the constraint)
- `active` defaults to `true`; setting to `false` prevents login but preserves the record

### Session

Represents an active login session tied to a user.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | text | PRIMARY KEY | Cryptographically random 64-character hex token (32 bytes) |
| user_id | integer | NOT NULL, REFERENCES users(id) ON DELETE CASCADE | The user who owns this session |
| created_at | timestamptz | NOT NULL, DEFAULT NOW() | When the session was created |
| expires_at | timestamptz | NOT NULL | When the session expires (7-day sliding window) |

**Indexes**: `idx_sessions_user_id` on `user_id`, `idx_sessions_expires_at` on `expires_at`.

**Validation Rules**:
- `id` is generated via `crypto/rand` (32 bytes, hex-encoded)
- Session is valid only when `expires_at > NOW()` AND the associated user has `active = true`. The `GetSession` query enforces both conditions -- a deactivated user's existing sessions are effectively invalidated at the query level without needing a separate revocation step.
- `expires_at` is extended by 7 days on each authenticated request (sliding window)
- Cascading delete: when a user is deleted, all their sessions are removed
- Expired sessions are not automatically deleted by the database. A `DeleteExpiredSessions` query runs on server startup to clean up stale rows.

## Relationships

```
User (1) ──── (many) Session
  │                    │
  │ id ◄───────── user_id (FK, CASCADE)
  │                    │
  └── email (UNIQUE)   └── id (crypto-random token, used as cookie value)
```

- One user can have many concurrent sessions (multiple devices/browsers)
- Deleting a user cascades to delete all their sessions
- Sessions reference users via `user_id` foreign key

## State Transitions

### User Lifecycle

```
[Bootstrap from env vars] → Active (role=owner)
                               │
                     (admin sets active=false via DB)
                               │
                               ▼
                           Inactive (cannot log in)
```

No self-service user creation, deletion, or role changes in this version.

### Session Lifecycle

```
[Login success] → Active (expires_at = now + 7 days)
                     │
            ┌────────┼────────┐
            │        │        │
    (each request)   │   (no activity
     extends by      │    for 7 days)
      7 days         │        │
            │        │        ▼
            │    (user logs   Expired
            │     out)        (middleware
            │        │         rejects,
            └────────┤         redirects
                     │         to login)
                     ▼
                  Deleted
              (row removed
               from DB)
```
