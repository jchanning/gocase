# GoCaSE — Vibe Coding Playbook Review

**Review Date:** 2026-03-14
**Playbook:** "The Vibe Coding Playbook" by Rob Vugts (v2.0)
**Reviewer:** GitHub Copilot

---

## Executive Summary

GoCaSE already satisfies several of the playbook's foundational setup requirements:

- the Go module is correctly named `github.com/jchanning/gocase`
- `.cursorrules`, `.cursorignore`, and `.cursor/rules/` exist
- `docs/ARCHITECTURE.md`, `docs/API.md`, and `docs/IMPLEMENTATION_STATUS.md` exist
- GitHub Actions CI exists and runs vet, staticcheck, tests with race detection and coverage, Docker build, and lint
- the current test suite is green

The remaining deficiencies are **architectural and process-oriented**, not basic project setup. The biggest gaps are:

1. the documentation spine is drifting from reality
2. the higher-order AaC documents are missing
3. the test suite has no fitness-function layer
4. handler/repository coupling makes strict TDD harder than it should be
5. repo hygiene and authoritative docs still have stale data

**Severity classification:**

| Severity | Count | Impact |
|----------|-------|--------|
| Critical | 1 | Living context has already become untrustworthy |
| High | 4 | Core AaC layers missing; tests not enforcing architecture |
| Medium | 4 | Thin coverage in critical areas; coupling hurts TDD |
| Low | 3 | Documentation and repo hygiene drift |

---

## Issues Found

### CRITICAL

#### C1 — Living docs are already stale and contradictory
**Files:** `docs/PLAYBOOK_REVIEW.md`, `docs/IMPLEMENTATION_STATUS.md`
**Problem:** The project now has `.cursorrules`, `.cursorignore`, `.cursor/rules/`, `docs/`, a fixed module name, and a CI workflow, but this review file and the status log still claim those items are missing or pending. Once the living docs diverge from reality, the playbook's "self-healing context" breaks down and future AI sessions start from false premises.
**Playbook Reference:** Chapter 4 — "The Living Context"; Pillar 3 — disciplined workflow.
**Fix:** Refresh the review and status docs immediately whenever setup work lands. Treat stale context as a blocking defect.
**Effort:** Low, ongoing.

---

### HIGH

#### H1 — Blueprint and Master Plan are still missing
**Missing files:** `docs/BLUEPRINT.md`, `docs/MASTER_PLAN.md`
**Problem:** The project has architecture overview docs, but not the constitutional design artifacts the playbook expects. There is no single document that captures product philosophy, core design principles, target users, non-functional priorities, and the architectural thesis that should govern future changes.
**Playbook Reference:** Chapters 8-10 — Blueprint and Master Plan.
**Fix:** Create both documents and make them the authoritative upstream source for future specs and ADRs.
**Effort:** Medium.

#### H2 — Non-Goals are not explicitly documented
**Missing file:** `docs/NON_GOALS.md`
**Problem:** The codebase has implied product boundaries, but the exclusions are not explicit. That leaves room for AI and human contributors to propose features that cut against the intended product direction, such as external OAuth, SMTP-based reminders, mobile apps, real-time collaboration, or SPA rewrites.
**Playbook Reference:** Chapter 10 — Non-Goals document.
**Fix:** Create a durable non-goals list with forbidden, deferred, and restricted capabilities.
**Effort:** Low-medium.

#### H3 — Domain specifications and invariants are missing
**Missing file:** `docs/DOMAIN_SPEC.md`
**Problem:** `schema.sql` describes structure, but the business invariants still live mostly in code and developer assumptions. Examples include publication rules, assignment authority, and attempt lifecycle constraints. Without an explicit domain contract, future features can violate business rules while still compiling and passing feature tests.
**Playbook Reference:** Chapter 9 — Domain Specifications.
**Fix:** Create a glossary, invariants, and state-machine sections for the core flows.
**Effort:** Medium.

#### H4 — Feature specs are still prompts, not contracts
**Directory:** `spec/`
**Problem:** The `*.agent.md` files capture request intent, not frozen implementation contracts. They are useful historical prompts, but they are not a reliable substitute for `spec.md`-style artifacts with explicit goals, interfaces, constraints, and acceptance tests.
**Playbook Reference:** Chapters 5 and 10 — Spec-Driven Development and Fractal Specifications.
**Fix:** Add a reusable feature spec template and begin replacing prompt-only artifacts for major flows.
**Effort:** Medium.

---

### MEDIUM

#### M1 — No fitness-function tests enforce architecture
**Problem:** The project has feature tests, but no architectural tests for route allowlists, handler layering, template render safety as a standard, or documented invariant enforcement. The code can stay green while architectural quality declines.
**Playbook Reference:** Chapter 9 — Fitness Functions.
**Fix:** Add an `internal/fitness/` package with tests for layer boundaries, public route policies, and critical business invariants.
**Effort:** High.

#### M2 — Test surface is still thin in critical runtime paths
**Current state:** the suite passes, but coverage is concentrated in auth middleware, parsing helpers, LLM parsing, storage helpers, and a few handler tests. There are still no repository tests, no tests for `dashboard_handler.go`, no tests for `teacher_handler.go`, and only partial admin/test flow coverage.
**Playbook Reference:** Chapter 6 — TDD.
**Fix:** Add repository-level tests and integration-style handler tests around the highest-value flows.
**Effort:** High.

#### M3 — Handlers depend on concrete repositories, which works against easy TDD
**Files:** `internal/handlers/*.go`, `internal/repository/*.go`
**Problem:** Handlers take concrete repository structs instead of narrow interfaces. That increases coupling and makes cheap, isolated tests harder than they should be. The playbook expects the implementation loop to be highly testable and replaceable; this design makes that more awkward.
**Playbook Reference:** Golden Trinity; TDD discipline.
**Fix:** Introduce small per-handler interfaces for the repository behavior each handler actually needs.
**Effort:** Medium-high.

#### M4 — Static docs still contradict the codebase
**Files:** `README.md`, `CHANGELOG.md`, `docs/IMPLEMENTATION_STATUS.md`
**Problem:** README still claims Go 1.22+, the standards list is incomplete, and CHANGELOG still references `my-app/internal/validation`. Those are small errors, but they degrade trust in the docs and create confusion for future sessions.
**Playbook Reference:** Pillar 3 — explicit rules and documentation.
**Fix:** Keep root docs aligned with the actual code and tooling in the same change set that modifies behavior.
**Effort:** Low.

---

### LOW

#### L1 — Repo hygiene around generated/user content is still weak
**Files:** `.gitignore`, `uploads/`, `.playwright-mcp/`
**Problem:** The AI context filter excludes uploads and screenshots, but source control boundaries are still loose. User-generated/reference material under `uploads/` lives inside the repo tree, and `.gitignore` does not currently reflect all of those boundaries.
**Playbook Reference:** Chapter 2 — clean context and clean project boundaries.
**Fix:** Strengthen `.gitignore` and keep user content outside source-controlled areas when practical.
**Effort:** Low.

#### L2 — `.env.example` still lacks relationship guidance
**File:** `.env.example`
**Problem:** Local and production env templates exist, but the relationship between `.env.example` and `deployment/env.production.example` is not explained.
**Playbook Reference:** Pillar 3 — explicit documentation.
**Fix:** Document which file is for local development versus production deployment.
**Effort:** Trivial.

#### L3 — Session volatility is documented but not yet promoted to a managed backlog item
**Files:** `docs/ARCHITECTURE.md`, `docs/IMPLEMENTATION_STATUS.md`
**Problem:** In-memory sessions are an acceptable early tradeoff, but they should remain visible as an explicit deferred architectural constraint rather than a buried note.
**Playbook Reference:** ADR discipline and operational rigor.
**Fix:** Keep session persistence in the tracked backlog until an explicit decision is made to retain it for v1.
**Effort:** Low.

---

## Resolution Plan

### Phase 1 — Documentation Truthfulness and AaC Baseline

| # | Task | Status |
|---|------|--------|
| 1 | Refresh stale living docs so they match the repo | Completed |
| 2 | Create `docs/BLUEPRINT.md` | Completed |
| 3 | Create `docs/MASTER_PLAN.md` | Completed |
| 4 | Create `docs/NON_GOALS.md` | Completed |
| 5 | Create `docs/DOMAIN_SPEC.md` | Completed |
| 6 | Add `spec/FEATURE_SPEC_TEMPLATE.md` | Completed |
| 7 | Fix root documentation drift (`README.md`, `CHANGELOG.md`, `.env.example`) | Completed |

### Phase 2 — Testability and Guardrails

| # | Task | Outcome |
|---|------|---------|
| 8 | Introduce small repository interfaces at handler boundaries | Easier unit tests and stricter TDD loop |
| 9 | Add repository tests | Validate SQL behavior and edge cases |
| 10 | Add missing handler tests | Cover dashboard, teacher, and admin critical flows |
| 11 | Add architectural fitness tests | Enforce layering, route policies, and invariants |

### Phase 3 — Cleanup and Consolidation

| # | Task | Outcome |
|---|------|---------|
| 12 | Consolidate static implementation docs into `docs/IMPLEMENTATION_STATUS.md` | Single authoritative log |
| 13 | Tighten repo hygiene for uploads/screenshots/artifacts | Cleaner source boundaries |
| 14 | Backfill proper specs for key flows | Authentication, test lifecycle, assignment lifecycle |

---

## End-of-Session Discipline

At the end of each coding session:

```
Review the work we just did. Update docs/IMPLEMENTATION_STATUS.md to check off
completed tasks, add newly discovered technical debt to Known Issues, and list
the next logical steps.
```
