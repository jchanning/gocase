# GoCaSE — Vibe Coding Playbook Review

**Review Date:** 2026-03-14
**Playbook:** "The Vibe Coding Playbook" by Rob Vugts (v2.0)
**Reviewer:** GitHub Copilot

---

## Executive Summary

GoCaSE is a well-functioning application with a solid domain model, clean Go architecture, and a working
deployment pipeline. However, measured against the playbook's Architecture-as-Code (AaC) framework and
Golden Trinity (SDD → TDD → Vibe Coding) methodology, it has several significant gaps. These gaps don't
affect the current running application but **will compound as the project grows** — each new AI session
lacks persistent context about design decisions, and there are no automated guardrails to catch architectural
drift.

**Severity classification:**

| Severity | Count | Impact |
|----------|-------|--------|
| Critical | 3 | AI sessions have no context; module naming is broken |
| High | 5 | AaC scaffolding absent; no living status doc |
| Medium | 4 | Test coverage thin; no fitness functions |
| Low | 2 | Minor quality items |

---

## Issues Found

### CRITICAL

#### C1 — Go module named `my-app` (placeholder not replaced)
**File:** `go.mod`
**Problem:** The module is named `my-app` — a scaffolding placeholder. This name bleeds into every import
path across the entire codebase (`my-app/internal/auth`, `my-app/internal/handlers`, etc.) and is
misleading for any developer or AI agent reading the code. Every import in every file contains this name.
**Playbook Reference:** Pillar 1 (Architect's Mindset — clear, intentional design choices).
**Fix:** Rename module to `github.com/jchanning/gocase` and update all import paths.
**Effort:** Medium (automated find-replace across ~15 files).

#### C2 — No `.cursorrules` (Global Constitution missing)
**File:** None — does not exist.
**Problem:** Every AI session starts with zero knowledge of the project's tech stack, conventions, and
non-negotiable constraints. The AI must re-learn preferences each session, leading to inconsistent code
style, duplicate logic, and hallucinated APIs. Without this file, every AI prompt is more expensive and
less accurate than it needs to be.
**Playbook Reference:** Chapter 4 — "The 'Living' Context", Pillar 1 Tip #9, Tip #19.
**Fix:** Create `.cursorrules` with tech stack, behavioral guidelines, and architectural constraints.
**Effort:** Low (one-time creation).

#### C3 — No `.cursorignore` (AI wastes tokens on irrelevant files)
**File:** None — does not exist.
**Problem:** Without this file, the AI indexes `go.sum`, compiled binaries (`.exe` files in root),
`uploads/` directory, `sample_tests/`, `.playwright-mcp/` screenshots, and `assets/` — thousands of
tokens of noise that degrade AI accuracy and increase cost.
**Playbook Reference:** Chapter 2 — "The `.cursorignore` File".
**Fix:** Create `.cursorignore` to exclude binaries, lock files, uploads, and generated assets.
**Effort:** Low (one-time creation).

---

### HIGH

#### H1 — No `docs/` directory with the three living documents
**Missing files:** `docs/ARCHITECTURE.md`, `docs/API.md`, `docs/IMPLEMENTATION_STATUS.md`
**Problem:** The playbook identifies these three as the "killer" context files. While the project has
`IMPLEMENTATION.md` and `FEATURES_IMPLEMENTED.md`, these are static snapshots frozen at feature-completion
time — not living documents updated each session. The AI has no map of where things currently stand, what
is broken, or what comes next.
**Playbook Reference:** Chapter 4 "The Documentation Strategy (My Personal Best Practice)".
**Fix:** Create `docs/` directory with all three files. `IMPLEMENTATION_STATUS.md` must become the
session-end update target.
**Effort:** Medium (initial creation; ongoing discipline required).

#### H2 — No Architectural Blueprint or Master Plan
**Missing files:** `docs/BLUEPRINT.md`, `docs/MASTER_PLAN.md`
**Problem:** The project has no document capturing the foundational vision, design philosophy, or
architectural decisions (ADRs). When new features are requested, the AI has no "North Star" to align
against — it derives intent from context alone, which leads to feature drift. There is also no C4
component model showing how the system pieces fit together.
**Playbook Reference:** Chapter 8-9 — Architecture-as-Code, Document Hierarchy.
**Fix:** Create Blueprint and Master Plan documents.
**Effort:** Medium.

#### H3 — No Non-Goals document
**Missing files:** `docs/NON_GOALS.md`
**Problem:** Without explicit scope exclusions, every new AI session may propose features outside the
product's intended scope. Examples of likely Non-Goals that have never been written down: real-time
collaboration, email notifications via SMTP, mobile native apps, external OAuth providers, etc.
**Playbook Reference:** Chapter 9 — "NON-GOALS (Explicit Scope Exclusions)".
**Fix:** Create `docs/NON_GOALS.md` with explicit exclusions across all feature categories.
**Effort:** Low-medium.

#### H4 — No Domain Specifications document
**Missing files:** `docs/DOMAIN_SPEC.md`
**Problem:** The database schema (`schema.sql`) captures the data model but not the business rules or
invariants. For example: "A test must have at least one question to be published", "A student cannot
retake a test they are currently attempting", "Assignments can only be created by teachers or admins" —
none of these invariants are written down anywhere. The AI must infer them from code, leading to subtle
violations.
**Playbook Reference:** Chapter 9 — "DOMAIN SPECIFICATIONS (Executable Technical Truth)".
**Fix:** Create `docs/DOMAIN_SPEC.md` with glossary, invariants, and state machine tables.
**Effort:** Medium.

#### H5 — Spec files are AI prompts, not contracts
**Directory:** `spec/`
**Problem:** The files in `spec/` are feature request prompts formatted as `.agent.md` — instructions
*to* the AI about what to build. The playbook distinguishes between prompts (inputs) and specs (contracts).
A proper `spec.md` captures the *outcome* of planning: data models, interface definitions, edge cases,
and constraints. The existing files tell the AI what to do but don't freeze the design decisions that
result from the conversation.
**Playbook Reference:** Chapter 5 — "The `spec.md` Artifact".
**Fix:** Create a `spec/FEATURE_SPEC_TEMPLATE.md` and retrospectively create proper spec documents for
key domain areas (authentication, test-taking flow, assignment lifecycle).
**Effort:** Medium.

---

### MEDIUM

#### M1 — Fitness functions absent (no architectural enforcement)
**Problem:** There are zero tests that validate architectural constraints — things like: "no handler
should directly access the database pool (must go through repository)", "all templates must render
without panic", "all routes must require authentication except the public allowlist". As the codebase
grows, architectural violations become invisible until runtime.
**Playbook Reference:** Chapter 9 — "FITNESS FUNCTIONS (Automated Enforcement)".
**Fix:** Add an `internal/fitness/` package with constraint-validation tests.
**Effort:** High (ongoing investment).

#### M2 — Test coverage is thin across critical paths
**Current state:** ~30 tests across 8 files. Missing coverage:
- `repository/` package — zero tests (all DB queries untested)
- `handlers/dashboard_handler.go` — no tests
- `handlers/admin_handler.go` — partial (only notes feature tested)
- `handlers/teacher_handler.go` — no tests
- `models/` — no validation tests
- Full request lifecycle (routing → handler → repository → response) — no integration tests
**Playbook Reference:** Chapter 6 — TDD, "Hallucination Control".
**Fix:** Add repository mock layer; add handler tests for all critical paths; add model validation tests.
**Effort:** High (ongoing investment).

#### M3 — No CI/CD pipeline defined
**Current state:** `README.md` references a CI badge (`ci.yml`) but no `.github/workflows/` directory
appears to exist locally (or tests are not run in CI). The deploy script runs manually from a developer's
machine.
**Playbook Reference:** Pillar 3 — "Bake in security and quality checks"; Chapter 8 — "Automated
Validation: Architectural rules are enforced by fitness functions that run on every commit."
**Fix:** Create `.github/workflows/ci.yml` to run `go test ./...` and `go vet` on every push and PR.
**Effort:** Low-medium.

#### M4 — No `.cursor/rules/` specialist rules directory
**Problem:** No language-specific or workflow-specific rules exist. The AI has no guidance on Go idioms,
template conventions, or testing patterns specific to this project.
**Playbook Reference:** Chapter 4 — "The Specialist Rules: `.cursor/rules/`".
**Fix:** Create `.cursor/rules/go.mdc`, `.cursor/rules/testing.mdc`, `.cursor/rules/templates.mdc`.
**Effort:** Low.

---

### LOW

#### L1 — Static documentation files overlap and are not maintained
**Files:** `IMPLEMENTATION.md`, `IMPLEMENTATION_COMPLETE.md`, `FEATURES_IMPLEMENTED.md`, `CHANGELOG.md`
**Problem:** Four separate files track implementation history, creating confusion about which is
authoritative. These were created at feature-completion time and are never updated mid-session. The
`CHANGELOG.md` contains developer-facing change notes but is not in conventional format (no versions).
**Fix:** Consolidate into `docs/IMPLEMENTATION_STATUS.md` (living) + keep `CHANGELOG.md` (versioned
with semver since v0.1.0).
**Effort:** Low (migration + discipline).

#### L2 — No `.env.example` for local development (only `env.production.example`)
**File:** `.env.example` exists at root but `deployment/env.production.example` is the authoritative
production template. The relationship between the two is undocumented.
**Fix:** Add a comment to `.env.example` clarifying its relationship to production config and what
values are required for local Docker Compose development.
**Effort:** Trivial.

---

## Resolution Plan

Tasks are ordered by priority. Each item is independently actionable.

### Phase 1 — Immediate (AI Context & Tooling) — Do these now

| # | Task | Files Created/Modified | Effort |
|---|------|------------------------|--------|
| 1 | Create `.cursorrules` | `.cursorrules` | 1h |
| 2 | Create `.cursorignore` | `.cursorignore` | 15m |
| 3 | Create `.cursor/rules/` specialist rules | `.cursor/rules/go.mdc`, `testing.mdc`, `templates.mdc` | 1h |
| 4 | Create `docs/IMPLEMENTATION_STATUS.md` | `docs/IMPLEMENTATION_STATUS.md` | 1h |
| 5 | Create `docs/ARCHITECTURE.md` | `docs/ARCHITECTURE.md` | 2h |
| 6 | Create `docs/API.md` | `docs/API.md` | 2h |
| 7 | Rename module from `my-app` to `github.com/jchanning/gocase` | `go.mod` + all `*.go` imports | 2h |

### Phase 2 — Architecture Documentation

| # | Task | Files Created/Modified | Effort |
|---|------|------------------------|--------|
| 8 | Create `docs/BLUEPRINT.md` | `docs/BLUEPRINT.md` | 2h |
| 9 | Create `docs/MASTER_PLAN.md` | `docs/MASTER_PLAN.md` | 3h |
| 10 | Create `docs/NON_GOALS.md` | `docs/NON_GOALS.md` | 1h |
| 11 | Create `docs/DOMAIN_SPEC.md` | `docs/DOMAIN_SPEC.md` | 3h |
| 12 | Create CI pipeline | `.github/workflows/ci.yml` | 1h |

### Phase 3 — Test Coverage & Fitness Functions

| # | Task | Files Created/Modified | Effort |
|---|------|------------------------|--------|
| 13 | Add repository layer tests (with DB mocking) | `internal/repository/*_test.go` | 4h |
| 14 | Add handler integration tests for critical paths | `internal/handlers/*_test.go` | 4h |
| 15 | Create fitness function tests | `internal/fitness/architecture_test.go` | 3h |
| 16 | Add model/validation unit tests | `internal/models/models_test.go`, `internal/validation/*_test.go` | 2h |

### Phase 4 — Cleanup

| # | Task | Effort |
|---|------|--------|
| 17 | Consolidate static docs → `docs/IMPLEMENTATION_STATUS.md` | 1h |
| 18 | Fix `.env.example` relationship documentation | 15m |
| 19 | Retroactively create proper spec.md for core domain flows | 3h |

---

## Quick Wins (Implemented as Part of This Review)

The following were created immediately as part of this review:

- ✅ `docs/PLAYBOOK_REVIEW.md` — this document
- ✅ `.cursorrules` — Global Constitution
- ✅ `.cursorignore` — AI context filter
- ✅ `.cursor/rules/go.mdc` — Go specialist rules
- ✅ `.cursor/rules/testing.mdc` — Testing specialist rules
- ✅ `docs/ARCHITECTURE.md` — Architecture overview
- ✅ `docs/API.md` — Route/API documentation
- ✅ `docs/IMPLEMENTATION_STATUS.md` — Living status document

---

## End-of-Session Discipline (Ongoing)

Per the playbook (Chapter 4), at the end of each coding session run:

```
Review the work we just did. Update docs/IMPLEMENTATION_STATUS.md to check off
completed tasks, add any new technical debt discovered to the Known Issues section,
and list the next logical steps.
```
