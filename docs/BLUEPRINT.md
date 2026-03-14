# GoCaSE Blueprint

**Version:** 0.1.0
**Status:** Draft
**Last Updated:** 2026-03-14

---

## 1. Core Problem Statement

Educational test delivery is often split across too many disconnected tools: content authoring,
assignment management, assessment delivery, and review analytics all live in different systems. GoCaSE
exists to provide a focused, low-friction assessment platform where teachers can create or upload tests,
assign them to students, and students can complete and review them with immediate feedback.

---

## 2. Product Thesis

GoCaSE should feel like a practical, dependable assessment workspace rather than a noisy LMS.
The product should prioritize:

- clarity over feature sprawl
- low operational complexity over architectural novelty
- deterministic core behavior over opaque automation
- server-rendered simplicity over frontend-framework complexity
- teacher control and student feedback over administrative ceremony

---

## 3. Target Users

### Students
Need a simple place to:
- discover assigned or available tests
- complete timed assessments
- understand their results immediately
- review answers and explanations
- receive sensible next-test recommendations

### Teachers
Need a practical workflow to:
- create, edit, and publish tests
- upload tests in bulk
- assign tests to selected students
- monitor attempts and completion status
- reuse notes and AI-generated question suggestions safely

### Admins
Need centralized control to:
- manage users and roles
- manage subjects/tests globally
- use AI-assisted generation features
- oversee the system without maintaining a separate admin product

---

## 4. Governing Product Laws

1. **Assessment First**
   The system exists to create, assign, take, and review tests. Features that do not clearly improve that loop are suspect.

2. **Deterministic Core**
   Test delivery, scoring, assignment state, and authorization must remain fully deterministic and auditable.

3. **AI at the Edge, Not the Core**
   AI may propose test content from notes, but it must not own scoring, authorization, or persisted state transitions.

4. **Simple Operations Matter**
   The product should remain deployable and operable on a single small Oracle Cloud instance without a large operational footprint.

5. **Server-Rendered by Default**
   The UI should continue using Go templates and HTMX unless a future ADR explicitly changes that.

---

## 5. Design Principles

- **Calm UI**: no unnecessary noise, popups, or gratuitous motion
- **Immediate feedback**: where pedagogy allows, students should see what happened and why
- **Visible state**: published/unpublished, pending/overdue/completed, pass/fail should always be legible
- **Low-click workflows**: frequent teacher/admin actions should be compact and obvious
- **Accessible forms**: labels, clear buttons, predictable navigation, and keyboard-friendly controls

---

## 6. Market Positioning

GoCaSE is not trying to be a full learning management system. It is a focused online assessment tool for
schools, tutors, and small educational teams that want a lightweight test platform with practical test
authoring and review workflows.

It competes by being:
- simpler to operate than a full LMS
- easier to extend than a closed SaaS quiz tool
- more teacher-centric than generic exam engines

---

## 7. Technical Strategy

- Go monolith with server-side rendering
- PostgreSQL as the single source of persisted truth
- HTMX for incremental UI interactions
- OCI GenAI only for bounded content-generation use cases
- Docker-based deployment to Oracle Cloud ARM64
- build-time version injection for release traceability

---

## 8. Non-Functional Priorities

### High Priority
- correctness of scoring and assignment state
- secure role-based authorization
- reliable deployment and reproducible builds
- maintainable codebase with small operational surface area

### Medium Priority
- fast page loads and simple HTMX interactions
- test coverage for critical flows
- documentation that survives across AI sessions

### Lower Priority for v0.x
- horizontal scaling
- distributed session storage
- external integrations beyond OCI GenAI

---

## 9. Architectural Direction

The authoritative implementation direction for v0.x is:
- monolith, not microservices
- repository-backed persistence, not ORM-heavy abstraction
- explicit handler/repository separation
- AI-assisted content generation, not AI-owned workflow logic
- production-safe simplicity over speculative extensibility

---

## 10. Relationship to Other Documents

This Blueprint is the constitutional product and design document.
It is upstream of:
- `docs/MASTER_PLAN.md`
- `docs/NON_GOALS.md`
- `docs/DOMAIN_SPEC.md`
- future feature specs under `spec/`

Changes to this document should be rare and intentional.