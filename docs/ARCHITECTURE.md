# GoCaSE — Architecture Overview

**Version:** 0.1.0
**Last Updated:** 2026-03-14

---

## System Overview

GoCaSE is a server-side rendered web application for delivering multiple-choice assessments to students
at various educational levels (Primary, Secondary, GCSE, A-Level). It is a monolithic Go application
with a PostgreSQL database, deployed as a Docker container behind an nginx reverse proxy.

```
Browser ──HTTPS──► nginx (443) ──HTTP──► gocase-app:8080 ──► PostgreSQL:5432
                    (TLS termination)    (Go/Chi server)      (pgx/v5)
```

---

## C4 Component Model

### Level 1 — System Context

```
[Student]  ──uses──►  [GoCaSE Web App]  ──reads/writes──►  [PostgreSQL DB]
[Teacher]  ──uses──►  [GoCaSE Web App]
[Admin]    ──uses──►  [GoCaSE Web App]  ──calls──►  [OCI GenAI (LLM)]
```

### Level 2 — Container Diagram

```
┌─────────────────────────────────────────────────────────┐
│  Oracle Cloud ARM64 VM                                   │
│                                                          │
│  ┌──────────┐    ┌────────────────────┐                 │
│  │  nginx   │    │  gocase-app        │                 │
│  │ (port 80 │───►│  (port 8080)       │                 │
│  │  & 443)  │    │  Go 1.25 / Chi v5  │                 │
│  └──────────┘    └─────────┬──────────┘                 │
│                            │ pgx/v5                      │
│                  ┌─────────▼──────────┐                 │
│                  │  gocase-db         │                 │
│                  │  PostgreSQL 14     │                 │
│                  │  (port 5432)       │                 │
│                  └────────────────────┘                 │
└─────────────────────────────────────────────────────────┘
                            │
                   OCI GenAI REST API
              (cohere.command-r-plus or Llama)
```

### Level 3 — Component Diagram (Go packages)

```
cmd/server/main.go
│
└── internal/server/server.go          ← Route definitions, middleware wiring
    ├── internal/auth/
    │   ├── session.go                 ← In-memory session store (cookie-based)
    │   └── middleware.go              ← RequireAuth, RequireRole middleware
    │
    ├── internal/handlers/
    │   ├── auth_handler.go            ← Login, Register, Logout
    │   ├── dashboard_handler.go       ← Student + Teacher dashboard
    │   ├── test_handler.go            ← Test listing, taking, submitting, results
    │   ├── teacher_handler.go         ← Test creation, editing, assignment
    │   └── admin_handler.go           ← Admin panel, user mgmt, LLM generation
    │
    ├── internal/repository/
    │   ├── user_repository.go         ← Users CRUD
    │   ├── test_repository.go         ← Tests, questions, answers, recommendations
    │   ├── attempt_repository.go      ← Test attempts, student answers, scoring
    │   └── assignment_repository.go   ← Teacher→student test assignments
    │
    ├── internal/models/models.go      ← Data structs (User, Test, Question, etc.)
    ├── internal/validation/           ← Input validation for test creation
    ├── internal/llm/                  ← OCI GenAI client + prompt builder
    ├── internal/docparse/             ← PDF/PPTX text extraction for notes
    ├── internal/storage/              ← Notes file upload/delete
    └── internal/version/version.go   ← Build-time version injection
```

---

## Data Model Summary

See `internal/database/schema.sql` for the authoritative schema.

**Core entities and relationships:**

```
users (id, email, role: student|teacher|admin)
  │
  ├── test_attempts (user_id → tests.id)   ← student takes a test
  │     └── student_answers               ← per-question responses
  │
  ├── user_achievements (user_id → achievements.id)
  ├── user_stats (points, streak, tests_completed)
  └── test_assignments (assigned_to → tests.id, assigned_by)

tests (id, title, subject_id, difficulty, exam_standard, published)
  ├── questions (test_id)
  │     └── answer_options (question_id, is_correct)
  ├── test_review_events               ← content review audit trail
  ├── test_feedback_issues             ← student-reported content issues
  └── subjects / topics
```

**Key constraints:**
- `difficulty` ∈ {Easy, Medium, Hard}
- `exam_standard` ∈ {Primary, Secondary, GCSE, IGCSE, A-Level}
- `role` ∈ {student, teacher, admin}
- `test_attempts.status` ∈ {in_progress, completed, abandoned}
- `test_assignments.status` ∈ {pending, completed, overdue}
- `tests.review_status` ∈ {draft, pending_review, approved, changes_requested}
- `test_feedback_issues.status` ∈ {open, in_review, resolved, dismissed}

---

## Authentication & Session Model

- Sessions are stored **in-memory** in a `sync.Map` keyed by a random cookie token
- Session lifetime: 24 hours (configurable in `auth/session.go`)
- Role enforcement: `auth.RequireRole(roles...)` middleware wraps protected route groups
- **No JWT. No OAuth. No external auth provider.**
- On server restart, all sessions are lost (users must log in again)

### Known Limitation
In-memory sessions do not survive container restarts or horizontal scaling. For v1.0, this is acceptable.
Future mitigation: persist sessions in PostgreSQL or Redis.

---

## LLM Integration

- Provider: Oracle Cloud Infrastructure (OCI) Generative AI Service
- Default model: `cohere.command-r-plus` (configurable via `OCI_GENAI_MODEL_ID` env var)
- Also supports: Meta Llama models via GENERIC API format
- Usage: Admin "Generate from Notes" feature — extracts text from uploaded PDF/PPTX and calls LLM to produce MCQ JSON
- Prompt template: `internal/llm/prompt.go`
- Auth: OCI API key (RSA private key + fingerprint) — key mounted as a file in Docker

---

## Deployment Architecture

```
Developer machine
│
├── git push → GitHub (github.com/jchanning/gocase)
│
└── .\deployment\deploy.ps1
      1. git archive HEAD → tar.gz (uses committed files only)
      2. scp tar.gz → Oracle Cloud VM
      3. SSH: docker compose build --build-arg BUILD_VERSION=<tag>
      4. SSH: docker compose up -d
      5. Health check: curl http://127.0.0.1:8081/
```

**Version tagging:** `git tag vX.Y.Z` before deploy → version appears in app footer.

---

## Architectural Decisions (ADRs)

### ADR-001: Server-side rendering with Go templates
**Decision:** Use Go `html/template` + HTMX for all UI, not a SPA framework.
**Rationale:** Reduces complexity (no build step, no JS bundle, no API versioning), Go templates are
safe against XSS by default, HTMX provides dynamic interactions without a full frontend framework.
**Consequences:** No rich client-side state; page transitions require HTMX or full page loads.

### ADR-002: In-memory session store
**Decision:** Store sessions in a `sync.Map` in-process rather than a database or Redis.
**Rationale:** Simplicity for initial deployment on a single-instance free tier VM.
**Consequences:** Sessions lost on restart; cannot scale horizontally without session migration.

### ADR-003: Repository pattern for all DB access
**Decision:** All PostgreSQL queries are encapsulated in `internal/repository/` structs.
**Rationale:** Enables future testing with interface mocks; separates SQL from HTTP handling logic.
**Consequences:** All handlers must depend on repository interfaces, never on `*pgxpool.Pool` directly.

### ADR-004: Monolithic deployment
**Decision:** Single Go binary + single PostgreSQL instance deployed on one VM.
**Rationale:** Free tier Oracle Cloud ARM64 instance; simplicity over scalability for initial version.
**Consequences:** No horizontal scaling; single point of failure. Acceptable for current user volume.

### ADR-005: No email service
**Decision:** No SMTP, no transactional email, no password-reset email.
**Rationale:** Avoids external service dependency and complexity for the initial version. Admin resets passwords manually.
**Consequences:** Password recovery requires admin intervention. Documented as a known limitation.

### ADR-006: Review-gated publication
**Decision:** Require explicit human approval before a test can be published.
**Rationale:** Test creation and AI generation are draft-producing workflows; publication is a separate quality gate.
**Consequences:** Staff edits can invalidate prior approval and publication must remain blocked until review is re-approved.
