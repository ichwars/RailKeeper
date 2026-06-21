# Security

RailKeeper is intended for local-first and small self-hosted installations. This document describes the implemented baseline and the remaining hardening work.

## Implemented

- no default credentials
- first-run setup gate
- Argon2id password hashing
- random session and CSRF tokens
- session token hashes stored in SQLite
- HTTP-only session cookie
- SameSite cookies
- current-user password change with automatic revocation of other sessions
- admin password resets revoke the affected user's active sessions
- optional secure cookies via `RAILKEEPER_COOKIE_SECURE=true`
- CSRF validation for API write requests
- role enforcement with Admin, Editor, Viewer and Messe roles
- Messe users are isolated from Viewer inventory routes and only receive the symbol read access needed by the exhibition function picker
- admin-only user management for local accounts and role assignment
- admin-only session review and targeted local session revocation
- persistent rate limiting for login and setup attempts
- audit logs for setup, login, logout, password, session, user, vehicle and exhibition-list changes
- structured security event review in the settings area
- application backups exclude local users, roles, sessions, rate limits, audit logs, password hashes, app settings and user settings
- backup validation warns about ignored authentication and settings tables instead of importing them
- configurable image and attachment upload size limits
- configurable allowed attachment extensions via `RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS`
- executable attachment extension blocklist
- server-side MIME detection for attachments
- attachment storage path confinement to the configured data directory
- external article, image and document fetches require public HTTP(S) URLs and reject private, loopback, link-local, multicast and unspecified targets before redirects and TCP dial
- static asset cache separation from API responses
- security headers: `nosniff`, `same-origin` referrer policy, frame blocking, CSP and permissions policy with camera limited to same-origin barcode scanning while microphone and geolocation stay disabled

## Operational Notes

- Put RailKeeper behind HTTPS before setting `RAILKEEPER_COOKIE_SECURE=true`.
- Keep the `/data` directory private and backed up.
- Do not expose the service directly to the internet without a reverse proxy and TLS.
- Article search fetches third-party pages and should be considered untrusted input; results are suggestions and require explicit user selection.

## Open Hardening Work

- optional public read-only vehicle pages with explicit per-item enablement
