# Feature Specification: Minimal Starter Template

**Feature Branch**: `001-starter-template`
**Created**: 2026-03-10
**Status**: Draft
**Input**: Build a minimal starter template application for yomauro with Go backend, Next.js frontend, session-based auth, login page, sidebar shell, and dashboard page showing "Yo Mauro" in large text.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - First-Time Setup and Login (Priority: P1)

An administrator sets up the application for the first time. They configure environment variables (database URL, admin email, admin password), run the setup command, and the system automatically creates the database tables and bootstraps the owner account. They then open the browser, see a login page, enter their credentials, and are redirected to a dashboard displaying "Yo Mauro" in large text.

**Why this priority**: This is the core end-to-end flow — without it, the application has no value. It validates that the full stack works: database, backend auth, frontend login, and protected dashboard.

**Independent Test**: Can be fully tested by running `make setup`, `make dev`, opening the browser, logging in, and seeing the dashboard. Delivers a working authenticated application.

**Acceptance Scenarios**:

1. **Given** a fresh database and valid ADMIN_EMAIL/ADMIN_PASSWORD environment variables, **When** the application starts for the first time, **Then** an owner account is automatically created with those credentials.
2. **Given** a running application with a bootstrapped owner, **When** the user navigates to the root URL, **Then** they are redirected to the login page.
3. **Given** the login page is displayed, **When** the user enters valid email and password and submits, **Then** they are redirected to the dashboard page showing "Yo Mauro" in large text.
4. **Given** the login page is displayed, **When** the user enters invalid credentials, **Then** an error message is displayed and they remain on the login page.

---

### User Story 2 - Session Persistence and Logout (Priority: P2)

A logged-in user closes their browser tab and returns later. Their session is still valid (within 7 days), so they land directly on the dashboard without re-entering credentials. When they click logout, their session is invalidated and they are returned to the login page.

**Why this priority**: Session management is essential for usability — users shouldn't need to log in every time they open the app. Logout is needed for security and multi-user machines.

**Independent Test**: Can be tested by logging in, closing the tab, reopening, verifying dashboard loads without login, then clicking logout and verifying redirect to login page.

**Acceptance Scenarios**:

1. **Given** a user with a valid session cookie, **When** they navigate to the dashboard, **Then** they see the dashboard without being asked to log in again.
2. **Given** a logged-in user on the dashboard, **When** they click the logout button in the sidebar, **Then** their session is invalidated and they are redirected to the login page.
3. **Given** a user whose session has expired (older than 7 days without activity), **When** they navigate to any protected page, **Then** they are redirected to the login page.

---

### User Story 3 - Mobile Navigation (Priority: P3)

A user accesses the application on a mobile device. Instead of a side sidebar, they see a bottom tab bar with Dashboard and Logout options. The dashboard content is responsive and readable on small screens.

**Why this priority**: Mobile support ensures the template is usable across devices. Lower priority because the primary use case is desktop, but the sidebar shell should be responsive from the start.

**Independent Test**: Can be tested by opening the app on a mobile viewport, verifying the bottom tab bar appears, navigating to dashboard, and logging out via the tab bar.

**Acceptance Scenarios**:

1. **Given** a logged-in user on a mobile device (viewport under 768px), **When** they view the dashboard, **Then** they see a bottom tab bar instead of a side sidebar.
2. **Given** a mobile user viewing the dashboard, **When** they tap the logout option in the bottom bar, **Then** they are logged out and redirected to the login page.

---

### Edge Cases

- What happens when the bootstrap runs but an owner already exists? The system skips owner creation silently.
- What happens when two simultaneous login attempts use the same credentials? Both succeed independently with separate session tokens.
- What happens when a user submits the login form with empty fields? The form shows a validation error without making a network request.
- What happens when the backend is unreachable from the frontend? The login form displays a connection error message.
- What happens when multiple browser tabs are open and the user logs out in one? Other tabs redirect to login on their next navigation or API call.
- What happens when the session cookie exists but the session has been deleted from the database? The middleware redirects to login and clears the stale cookie.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST bootstrap an owner account from environment variables on first startup when no active owners exist.
- **FR-002**: System MUST provide a login page that accepts email and password credentials.
- **FR-003**: System MUST authenticate users by comparing submitted passwords against stored hashes using a secure one-way hashing algorithm.
- **FR-004**: System MUST prevent timing-based user enumeration by performing constant-time password comparison even for non-existent accounts.
- **FR-005**: System MUST rate-limit login attempts to 5 per IP address per minute.
- **FR-006**: System MUST create secure session cookies (HttpOnly, Secure, SameSite=Lax) with a 7-day expiry on successful login.
- **FR-007**: System MUST extend the session expiry by 7 days on each authenticated request (sliding window).
- **FR-008**: System MUST validate session cookies on every protected page request by checking the session exists and has not expired.
- **FR-009**: System MUST cache validated user data in a short-lived cookie (5-minute TTL) to reduce repeated backend validation calls.
- **FR-010**: System MUST redirect unauthenticated users to the login page when they attempt to access any protected page.
- **FR-011**: System MUST invalidate the session (delete from storage and clear cookie) when the user logs out.
- **FR-012**: System MUST display a dashboard page showing "Yo Mauro" in large, prominent text after successful login.
- **FR-013**: System MUST provide a sidebar navigation shell with a Dashboard link and a Logout action.
- **FR-014**: System MUST display a bottom tab bar on mobile viewports (under 768px) instead of the sidebar.
- **FR-015**: System MUST provide all operational commands (server start, build, migrate, code generation, frontend dev/build, setup) through a single root Makefile.

### Key Entities

- **User**: Represents an authenticated account. Key attributes: email (unique identifier for login), display name (shown in UI), role (always "owner" in this version), active status (soft-delete flag).
- **Session**: Represents an active login session. Key attributes: cryptographically random token (used as cookie value), association with a user, creation time, expiration time (7-day sliding window).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new developer can go from cloning the repository to seeing the "Yo Mauro" dashboard in under 5 minutes using documented make commands.
- **SC-002**: The login flow completes (form submission to dashboard render) in under 2 seconds on a local development environment.
- **SC-003**: Invalid login attempts return a clear error message within 1 second without revealing whether the email exists.
- **SC-004**: Session persistence works across browser restarts — a user who logged in within the last 7 days is not asked to log in again.
- **SC-005**: The application renders correctly on viewports from 320px (mobile) to 1920px (desktop) with appropriate navigation patterns at each size.

## Assumptions

- PostgreSQL is available locally or via a connection string provided in environment variables.
- The developer has Go, Node.js, and Make installed on their machine.
- The application runs in development mode by default (relaxed validation, permissive CORS).
- Only one user (the bootstrapped owner) will exist in this minimal version; user management is out of scope.
- HTTPS is not required in development; the Secure cookie flag will be adapted based on environment.
