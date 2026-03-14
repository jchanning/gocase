# GoCaSE — Implementation Status

**Current Version:** 0.1.0
**Last Updated:** 2026-03-14
**Production URL:** https://gocase.fistraltech.com

> This is a **living document**. Update it at the end of every coding session:
> *"Review the work we just did. Update docs/IMPLEMENTATION_STATUS.md to check off completed tasks,
> add any new technical debt to Known Issues, and list next logical steps."*

---

## Completed Features ✅

### Authentication & Users
- [x] User registration (student self-registration; teachers/admins created by admin)
- [x] Cookie-based session login/logout
- [x] Role-based access control (student / teacher / admin)
- [x] Admin user management (create, update role, reset password, delete)
- [x] Registration prevents self-elevation to teacher/admin role

### Tests — Student Experience
- [x] Browse and filter published tests (subject, difficulty, exam standard)
- [x] Group tests by subject, difficulty, or exam standard
- [x] Take timed multiple-choice tests (timer displayed, auto-submit on expiry)
- [x] Per-question feedback during test (HTMX-driven)
- [x] Immediate results with score, pass/fail, and question breakdown
- [x] Explanations per question shown on results and review pages
- [x] Test review page (all questions with correct/incorrect highlighting)
- [x] Test history with filters

### Tests — Teacher/Admin Management
- [x] Upload tests via JSON file
- [x] Create tests manually (title, subject, difficulty, exam standard, questions)
- [x] Edit existing tests (metadata + questions)
- [x] Publish / unpublish tests
- [x] Preview tests as a student would see them
- [x] Delete tests
- [x] PDF export of tests (admin only)
- [x] Assign tests to specific students with due dates
- [x] Mark overdue assignments automatically

### LLM / AI Features
- [x] Upload PDF/PPTX notes file to a test
- [x] "Generate from Notes" — calls OCI GenAI to produce MCQ JSON from notes
- [x] Supports Cohere Command R+ and Meta Llama via OCI GenAI (GENERIC format)
- [x] Generated questions editable before saving

### Dashboard & Analytics
- [x] Student dashboard: pending assignments, recent results, stats
- [x] Teacher dashboard: class overview, test list, assignment list
- [x] Gamification: points, badges/achievements, streaks (user_stats, achievements tables)
- [x] Test recommendations after completion (next difficulty level in same subject)

### Infrastructure & Deployment
- [x] Docker + docker-compose.prod.yml for OCI ARM64 deployment
- [x] nginx reverse proxy with TLS (Let's Encrypt via certbot)
- [x] Automated deploy script (`deployment/deploy.ps1`)
- [x] Build-time version injection (`internal/version/version.go`)
- [x] Version `0.1.0` tagged and deployed

### UI/UX
- [x] Responsive design with Tailwind utility classes
- [x] Consistent design system (semantic tokens, reusable component classes)
- [x] HTMX for dynamic interactions (answer submission, delete confirmations)
- [x] Version number in footer

---

## In Progress 🚧

*(Nothing currently in progress — session ended cleanly)*

---

## Known Issues / Technical Debt 🐛

| ID | Area | Description | Severity |
|----|------|-------------|----------|
| TD-01 | Module | Go module rename completed: `github.com/jchanning/gocase` | DONE |
| TD-02 | Sessions | In-memory session store lost on container restart | Low (acceptable for v0.x) |
| TD-03 | Tests | No repository layer tests (all DB queries untested) | High |
| TD-04 | Tests | No handler tests for dashboard, teacher, or admin handlers | Medium |
| TD-05 | Fitness | No architectural fitness function tests | Medium |
| TD-06 | Docs | Core AaC docs created: `BLUEPRINT.md`, `MASTER_PLAN.md`, `NON_GOALS.md`, `DOMAIN_SPEC.md` | DONE |
| TD-07 | Specs | `spec/*.agent.md` files are prompts; contract-style specs still need backfilling | Medium |
| TD-08 | Email | No password reset via email (admin must reset manually) | Low |
| TD-09 | Scale | Single-instance deployment; no horizontal scaling | Low (v0.x) |
| TD-10 | Sessions | No session refresh on activity (24h hard expiry) | Low |
| TD-11 | PPTX | PPTX text extraction is limited; complex layouts may miss content | Low |
| TD-12 | Docs | Root docs drift was corrected; keep future docs synchronized in the same change set | DONE |

---

## Next Logical Steps 📋

Priority order based on current state:

1. **[High]** Deploy the accumulated fixes and playbook-alignment work to Oracle Cloud
2. **[Medium]** Expand architectural fitness tests beyond the initial guardrails
3. **[Low]** Consolidate `IMPLEMENTATION.md`, `IMPLEMENTATION_COMPLETE.md`, `FEATURES_IMPLEMENTED.md` into this file

---

## Session Log

### 2026-03-14 — v0.1.0 Release
**Work done:**
- Implemented all spec/12 features: teacher assignments, test recommendations, test groupBy, version display
- Fixed `parseTestFilters` defaulting published=true (was leaking unpublished tests)
- Set up versioning: `BUILD_VERSION` from git tag injected at Docker build time
- Tagged `v0.1.0` — deployed to production
- Applied Vibe Coding Playbook review: created `.cursorrules`, `.cursorignore`, `.cursor/rules/`, `docs/` structure

**Files changed (key):**
- `internal/handlers/test_handler.go` — published filter default fix
- `internal/repository/assignment_repository.go` — new (teacher assignments)
- `internal/repository/test_repository.go` — added `GetRecommendation`
- `views/teacher_assign.html` — new (assign test to students)
- `views/tests_list.html` — groupBy feature
- `views/test_results.html` — recommendation card
- `views/dashboard.html` — pending assignment cards with overdue status
- `deployment/deploy.ps1` — BUILD_VERSION from git tag
- `deployment/docker-compose.prod.yml` — BUILD_VERSION arg

### 2026-03-14 — Playbook Phase 1 Follow-up
**Work done:**
- Refreshed stale playbook review findings to match the real repo state
- Implemented the missing AaC document layer for Blueprint, Master Plan, Non-Goals, and Domain Spec
- Corrected root documentation drift and backlog priorities
- Added a reusable contract-style feature spec template

**Files changed (key):**
- `docs/PLAYBOOK_REVIEW.md` — replaced outdated findings with current-state review
- `docs/IMPLEMENTATION_STATUS.md` — updated backlog and next steps
- `docs/BLUEPRINT.md` — new
- `docs/MASTER_PLAN.md` — new
- `docs/NON_GOALS.md` — new
- `docs/DOMAIN_SPEC.md` — new
- `spec/FEATURE_SPEC_TEMPLATE.md` — new
- `README.md` — version and standards cleanup
- `CHANGELOG.md` — stale import path cleanup
- `.env.example` — clarified local vs production env usage

### 2026-03-14 — Tests Page Visibility Fix
**Work done:**
- Fixed `/tests` so students still default to published-only tests, while teachers and admins now see all created tests by default
- Added handler tests covering student, teacher, and admin default filter behavior

**Files changed (key):**
- `internal/handlers/test_handler.go` — role-aware default for published filter
- `internal/handlers/test_handler_filters_test.go` — added role-specific default filter coverage

### 2026-03-14 — Playbook Phase 2 Follow-up
**Work done:**
- Backfilled contract-style specs for authentication, test lifecycle, and assignment lifecycle
- Added the first `internal/fitness` architecture tests for handler-layer boundaries, protected route structure, and required docs

**Files changed (key):**
- `spec/authentication.spec.md` — new
- `spec/test-lifecycle.spec.md` — new
- `spec/assignment-lifecycle.spec.md` — new
- `internal/fitness/architecture_test.go` — new

### 2026-03-14 — Playbook Phase 3 Follow-up
**Work done:**
- Replaced concrete repository dependencies in `AuthHandler` and `DashboardHandler` with narrow interfaces
- Removed the unused `testRepo` dependency from `DashboardHandler`
- Added unit tests for auth success/error flows and dashboard data loading behavior
- Extended fitness tests so `auth_handler.go` and `dashboard_handler.go` cannot regress back to concrete repository types

**Files changed (key):**
- `internal/handlers/auth_handler.go` — interface boundary + register error fix
- `internal/handlers/dashboard_handler.go` — interface boundary + extracted dashboard data loader
- `internal/handlers/auth_handler_test.go` — expanded auth coverage
- `internal/handlers/dashboard_handler_test.go` — new
- `internal/fitness/architecture_test.go` — stronger boundary guard
- `internal/server/server.go` — constructor wiring updated

### 2026-03-14 — Playbook Phase 4 Follow-up
**Work done:**
- Replaced concrete repository dependencies in `TestHandler` with narrow interfaces
- Extracted submission, recommendation, and history-filter helper logic into directly testable units
- Added unit tests for history scoping, score calculation, stats updates, and recommendation gating
- Extended the fitness tests so `test_handler.go` cannot regress back to concrete repository types

**Files changed (key):**
- `internal/handlers/test_handler.go` — interface boundary + extracted helper logic
- `internal/handlers/test_handler_unit_test.go` — new
- `internal/fitness/architecture_test.go` — stronger boundary guard

### 2026-03-14 — Playbook Phase 5 Follow-up
**Work done:**
- Added `pgxmock`-backed repository tests for recommendation and assignment behavior
- Introduced a minimal repository DB interface so repositories can be unit tested without a live PostgreSQL instance

**Files changed (key):**
- `internal/repository/db.go` — new shared DB query interface
- `internal/repository/test_repository.go` — now depends on the minimal DB interface
- `internal/repository/assignment_repository.go` — now depends on the minimal DB interface
- `internal/repository/test_repository_test.go` — new
- `internal/repository/assignment_repository_test.go` — new

### 2026-03-14 — Playbook Phase 6 Follow-up
**Work done:**
- Replaced concrete repository dependencies in `TeacherHandler` and `AdminHandler` with narrow interfaces
- Refactored the shared test-upload helper to use an interface instead of a concrete repository
- Added staff-flow tests for publish/unpublish ownership, assignment validation/creation, subject creation, user creation, and role updates
- Extended fitness tests so `teacher_handler.go` and `admin_handler.go` cannot regress back to concrete repository types

**Files changed (key):**
- `internal/handlers/test_creation_helper.go` — interface-based upload persistence
- `internal/handlers/teacher_handler.go` — interface boundary + small ownership helper
- `internal/handlers/admin_handler.go` — interface boundary
- `internal/handlers/staff_handler_test.go` — new
- `internal/fitness/architecture_test.go` — stronger boundary guard

### 2026-03-14 — Playbook Phase 7 Follow-up
**Work done:**
- Added test-taking lifecycle handler coverage for start, answer submission, and test submission side effects
- Added admin document-generation handler coverage for unavailable/missing/invalid-upload paths
- Added admin notes-management handler coverage for remove and serve flows

**Files changed (key):**
- `internal/handlers/test_handler_flow_test.go` — new
- `internal/handlers/admin_handler_flow_test.go` — new
