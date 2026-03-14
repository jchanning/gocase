# GoCaSE Non-Goals

**Version:** 0.1.0
**Status:** Draft
**Last Updated:** 2026-03-14

This document defines capabilities that GoCaSE does not implement, should not drift into accidentally,
or should treat as explicitly deferred.

---

## Status Legend

- `FORBIDDEN` — not aligned with product direction for v0.x
- `DEFERRED` — potentially valid later, but explicitly not current scope
- `RESTRICTED` — allowed only within a tight boundary

---

## Product Scope Exclusions

### NG-001 — Full LMS feature set
**Status:** FORBIDDEN
GoCaSE is not a full learning management system with coursework, gradebooks, lesson plans, and messaging.

### NG-002 — Real-time collaboration on tests
**Status:** FORBIDDEN
No live multi-user editing, presence indicators, or collaborative authoring in v0.x.

### NG-003 — Native mobile apps
**Status:** DEFERRED
No iOS/Android native clients. Mobile support remains browser-based.

### NG-004 — Social/gamified engagement loops beyond current implementation
**Status:** RESTRICTED
Existing points/badges/streaks may remain, but no expansion into social feeds, competitive leaderboards,
or dopamine-driven mechanics without explicit product review.

---

## Architecture Exclusions

### NG-005 — SPA rewrite
**Status:** FORBIDDEN
Do not migrate the UI to React, Vue, Next.js, or another SPA framework without an ADR.

### NG-006 — Microservice split
**Status:** FORBIDDEN
Do not split the monolith into multiple services in v0.x.

### NG-007 — ORM migration
**Status:** FORBIDDEN
Do not replace explicit pgx/sql-style repositories with a heavy ORM.

### NG-008 — Event-driven rewrite
**Status:** DEFERRED
No message broker or async service mesh unless required by a future scaling ADR.

---

## Auth / Identity Exclusions

### NG-009 — OAuth / SSO providers
**Status:** DEFERRED
No Google, Microsoft, GitHub, or SAML auth in the current product.

### NG-010 — JWT-based auth architecture
**Status:** FORBIDDEN
Do not replace cookie-based sessions with JWT by default.

### NG-011 — Self-service password recovery via email
**Status:** DEFERRED
Password reset remains admin-managed until a full email and token lifecycle is designed.

---

## AI Boundary Exclusions

### NG-012 — AI-controlled scoring
**Status:** FORBIDDEN
AI must not determine correctness, scores, or pass/fail outcomes.

### NG-013 — AI direct mutation of persisted state
**Status:** FORBIDDEN
AI may propose question content, but it must not directly publish tests, assign tests, or change user state.

### NG-014 — Conversational tutor/chatbot product pivot
**Status:** DEFERRED
GoCaSE is not currently a tutoring chatbot or conversational learning assistant.

---

## Notification / Communication Exclusions

### NG-015 — Email reminder system
**Status:** DEFERRED
Assignments may have due dates and overdue status, but SMTP-based reminder infrastructure is not part of v0.x.

### NG-016 — Push notifications or SMS
**Status:** FORBIDDEN
No mobile push, browser push, or SMS reminders in current scope.

---

## Data / Integration Exclusions

### NG-017 — Public API platform
**Status:** DEFERRED
The system does not expose a public REST/GraphQL API for external integrators.

### NG-018 — Multi-tenant enterprise workspace model
**Status:** DEFERRED
The current product is not designed as a generalized multi-tenant SaaS platform.

### NG-019 — External object storage
**Status:** DEFERRED
No S3-style document storage until scale or compliance requirements justify it.

---

## Enforcement Rule

Features that fall into the FORBIDDEN set must not be implemented without an explicit ADR and an update to this document.