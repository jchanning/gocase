# GoCaSE Master Plan

**Version:** 0.1.0
**Status:** Draft
**Last Updated:** 2026-03-14

---

## 1. Architectural Thesis

GoCaSE is a server-rendered Go monolith optimized for clarity, deterministic assessment behavior,
and low operational complexity. The system keeps business-critical behavior inside a conventional
handler → repository → database flow and isolates AI usage to bounded content-generation workflows.

---

## 2. Technology Stack

| Layer | Technology | Rationale |
|------|------------|-----------|
| Application language | Go 1.25 | Simple deployment, strong standard library, fast server-side work |
| Router | Chi v5 | Minimal, idiomatic routing without framework weight |
| Database | PostgreSQL 14 + pgx/v5 | Explicit SQL, strong relational model, reliable on OCI |
| Templates | Go `html/template` | Safe SSR, low complexity, XSS-resistant by default |
| UI interactivity | HTMX + TailwindCSS | Lightweight progressive enhancement without SPA overhead |
| Auth | Cookie-based in-memory sessions | Simplicity for single-instance v0.x deployment |
| File handling | Local filesystem uploads | Operationally simple for current deployment model |
| AI integration | OCI Generative AI | Bounded note-to-question generation feature |
| Deployment | Docker + nginx on Oracle Cloud ARM64 | Cost-effective and reproducible deployment |

---

## 3. Container View

### Main Runtime Components
- nginx reverse proxy
- Go application container
- PostgreSQL container
- OCI GenAI as external service dependency

### Request Flow
1. User requests hit nginx over HTTPS
2. nginx proxies to the Go application
3. handlers coordinate request parsing and rendering
4. repositories execute SQL against PostgreSQL
5. templates render HTML responses

---

## 4. Component Boundaries

### `internal/handlers/`
Responsibilities:
- parse HTTP requests
- validate request parameters and form values
- call repository methods
- render templates or redirect

Must not:
- construct raw SQL
- own persistence details
- bypass auth/role rules

### `internal/repository/`
Responsibilities:
- own SQL queries
- map rows into models
- hide persistence details from handlers

Must not:
- render templates
- write HTTP responses
- embed UI logic

### `internal/auth/`
Responsibilities:
- session storage and lookup
- request middleware for auth/roles

### `internal/llm/`
Responsibilities:
- prompt construction
- OCI GenAI request formatting
- parsing generated test output

Must not:
- write directly to persistent domain state without explicit handler/admin action

---

## 5. Security Posture

### Core Controls
- role-based access control in middleware
- parameterized SQL via pgx
- server-rendered templates with escaped output
- cookie-based sessions
- admin-only AI generation workflow
- local key mount for OCI private key material

### Security Constraints
- no JWT unless introduced by ADR
- no external auth providers in v0.x
- no direct SQL string interpolation from user input
- no AI-generated content persisted without explicit save workflow

---

## 6. Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-01 | Release traceability | Every deploy exposes a visible version string |
| NFR-02 | Build reproducibility | Docker build must succeed from committed state |
| NFR-03 | Test gate | CI must run vet, lint, and tests before merge |
| NFR-04 | Operational simplicity | Single-instance deployment remains supported |
| NFR-05 | Authorization integrity | Student/teacher/admin boundaries must be enforced consistently |
| NFR-06 | Documentation continuity | Core docs must remain accurate across sessions |

---

## 7. ADR Index

### ADR-001
Use Go templates + HTMX instead of an SPA.

### ADR-002
Use in-memory cookie session storage for v0.x single-instance deployment.

### ADR-003
Use repository pattern for all SQL access.

### ADR-004
Keep deployment monolithic on Oracle Cloud ARM64.

### ADR-005
Use OCI GenAI only for bounded content-generation workflows.

Future ADRs should be added when:
- session persistence changes
- deployment topology changes
- frontend architecture changes
- external auth or messaging integrations are introduced

---

## 8. Readiness Gates for Future Features

A feature should not move to implementation until:
- it has a spec document or story-level contract
- any new persistence fields are reflected in schema and models
- authorization expectations are explicit
- edge cases are listed
- tests for the critical path are planned first

---

## 9. Planned Architecture Improvements

1. add domain specifications and invariants
2. add fitness-function tests
3. replace concrete repository dependencies in handlers with narrow interfaces
4. tighten repository and handler test coverage
5. clean up stale root-level documents and reduce duplicate sources of truth