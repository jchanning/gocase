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
| TD-01 | Module | Go module was named `my-app` — fixed to `github.com/jchanning/gocase` | ~~Medium~~ DONE |
| TD-02 | Sessions | In-memory session store lost on container restart | Low (acceptable for v0.x) |
| TD-03 | Tests | No repository layer tests (all DB queries untested) | High |
| TD-04 | Tests | No handler tests for dashboard, teacher, or admin handlers | Medium |
| TD-05 | CI/CD | No GitHub Actions CI pipeline (tests not auto-run on push) | Medium |
| TD-06 | Fitness | No architectural fitness function tests | Medium |
| TD-07 | Email | No password reset via email (admin must reset manually) | Low |
| TD-08 | Scale | Single-instance deployment; no horizontal scaling | Low (v0.x) |
| TD-09 | Sessions | No session refresh on activity (24h hard expiry) | Low |
| TD-10 | PPTX | PPTX text extraction is limited; complex layouts may miss content | Low |

---

## Next Logical Steps 📋

Priority order based on current state:

1. **[High]** Create GitHub Actions CI pipeline (`.github/workflows/ci.yml`)
   — Run `go test ./...` and `go vet ./...` on every push and PR
2. **[High]** Fix module name: rename `my-app` → `github.com/jchanning/gocase`
3. **[Medium]** Add repository layer tests using interface-based mocks
4. **[Medium]** Add handler integration tests for dashboard and teacher flows
5. **[Medium]** Create `docs/DOMAIN_SPEC.md` with invariants and state machines
6. **[Medium]** Create `docs/NON_GOALS.md` with explicit scope exclusions
7. **[Low]** Add architectural fitness function tests (`internal/fitness/`)
8. **[Low]** Consolidate `IMPLEMENTATION.md`, `IMPLEMENTATION_COMPLETE.md`, `FEATURES_IMPLEMENTED.md` into this file

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
