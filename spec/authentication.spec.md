# Authentication and Authorization Specification

## 1. Feature Name
Authentication and Authorization

## 2. Goal
Provide secure sign-in, session handling, role-based access control, and safe registration workflows for students, teachers, and admins.

## 3. Scope
- login with email and password
- registration for students
- admin-created teacher/admin accounts
- cookie-based sessions
- role-based route protection
- logout and session invalidation

## 4. Non-Scope
- OAuth / SSO
- JWT-based auth
- email-based password reset
- multi-factor authentication

## 5. Affected Users / Roles
- student
- teacher
- admin

## 6. Constraints
- sessions are cookie-based and server-stored
- passwords must be hashed with bcrypt
- elevated role creation requires admin authority
- handlers must not bypass middleware role checks
- no auth state may be derived from client-side claims alone

## 7. Data Model Impact
- `users`
- in-memory session store in `internal/auth/session.go`
- no DB-backed session table in v0.x

## 8. Interface / Route Impact
- `GET /login`
- `POST /login`
- `GET /register`
- `POST /register`
- `GET /logout`
- protected route groups in `internal/server/server.go`

## 9. Business Rules / Invariants
- only valid roles may be persisted
- non-admin users may not create teacher/admin accounts
- only authenticated users may access protected routes
- only users with allowed roles may access role-protected routes
- logout must invalidate the active session token

## 10. Edge Cases
- invalid email/password pair
- expired or missing session cookie
- user attempts self-elevation to teacher/admin
- malformed role input during registration

## 11. Acceptance Tests
1. Given valid student credentials, when login succeeds, then a session cookie is set and the user is redirected to `/dashboard`.
2. Given valid admin or teacher credentials, when login succeeds, then a session cookie is set and the user is redirected to `/admin`.
3. Given a non-admin trying to register a teacher/admin account, when the form is submitted, then the request is rejected with `403 Forbidden`.
4. Given a protected route and no valid session, when the route is requested, then the user is redirected or blocked by auth middleware.
5. Given a role-protected route and the wrong role, when the route is requested, then the user receives `403 Forbidden`.

## 12. TDD Plan
- extend `internal/auth/middleware_test.go`
- extend `internal/handlers/auth_handler_test.go`
- add negative-path tests around session expiry and elevated role creation
- keep middleware and handler tests separate

## 13. Documentation Updates
- `docs/IMPLEMENTATION_STATUS.md`
- `docs/API.md` if route behavior changes
- `docs/DOMAIN_SPEC.md` if auth invariants change

## 14. Rollout Notes
- any future move away from in-memory sessions requires an ADR
- any external auth provider requires updates to `docs/NON_GOALS.md` and `docs/MASTER_PLAN.md`