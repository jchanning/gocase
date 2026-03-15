# GoCaSE Domain Specification

**Version:** 0.1.0
**Status:** Draft
**Last Updated:** 2026-03-14

This document captures the core domain language, invariants, and lifecycle rules that must remain true
regardless of implementation detail.

---

## 1. Glossary

### User
A person with an account in the system.
Roles are `student`, `teacher`, or `admin`.

### Subject
A top-level academic category used to organize tests.

### Topic
A sub-category within a subject.

### Test
A composed assessment containing metadata, questions, answer options, publication state, and optional notes linkage.

### Question
A single assessment item belonging to a test.

### Answer Option
One of exactly four candidate answers for a question.

### Test Attempt
A student's in-progress or completed execution of a test.

### Student Answer
A persisted answer to a specific question within an attempt.

### Test Assignment
A teacher/admin-issued requirement that a specific student complete a specific test by a due date.

### Recommendation
A suggested next test shown after completion, based on subject and difficulty progression.

### Review Status
The publication governance state for a test. Valid values are `draft`, `pending_review`, `approved`, and `changes_requested`.

### Feedback Issue
A student-reported concern tied to a completed attempt, test, and question so teachers/admins can review and resolve content quality problems.

---

## 2. Core Invariants

### INV-001 — Role integrity
A user must always have exactly one valid role from the allowed role set.

### INV-002 — Question option cardinality
Each question must have exactly four answer options.

### INV-003 — Single correct answer
Each question must have exactly one correct answer option.

### INV-004 — Published test visibility
Students may only discover and start tests that are published, unless a handler explicitly uses elevated permissions.

### INV-005 — Assignment authority
Only a teacher or admin may create a test assignment.

### INV-006 — Assignment target validity
A test assignment must target a valid student user and a valid test.

### INV-007 — Attempt ownership
A student may only view, answer, submit, or review their own attempts.

### INV-008 — Deterministic scoring
Scores and correctness must be computed from persisted answer-option truth, not AI output.

### INV-009 — Recommendation boundary
Recommendations may suggest a next test, but must not modify assignment state or persistence.

### INV-010 — AI boundary
AI may generate candidate test content from notes, but it may not directly publish tests, assign tests, or alter attempt results.

### INV-011 — Review gate before publication
A test must not be published unless its current review status is `approved`.

### INV-012 — Review audit integrity
Submitting, approving, or requesting changes on a test must leave an audit event, and approval decisions must record reviewer identity and timestamp.

### INV-013 — Feedback ownership
A student may only report an issue for their own completed attempt and only against questions that belong to that attempt's test.

### INV-014 — Feedback moderation authority
Only a teacher or admin may change feedback issue status or record a moderation response.

---

## 3. State Machines

### Test Publication State

| State | Event | Next State |
|------|-------|------------|
| draft/unpublished | publish | published |
| published | unpublish | draft/unpublished |
| draft/unpublished | delete | deleted |
| published | delete | deleted |

### Assignment State

| State | Event | Next State |
|------|-------|------------|
| pending | due date passes without completion | overdue |
| pending | assigned student completes matching test flow | completed |
| overdue | matching test completed | completed |

### Attempt State

| State | Event | Next State |
|------|-------|------------|
| in_progress | student submits test | completed |
| in_progress | timer expiry auto-submit | completed |
| in_progress | explicit abandonment logic (future) | abandoned |

---

## 4. Rules for Publication

A test should not be published unless:
- it has valid metadata
- it has at least one question
- each question has four options
- each question has exactly one correct answer

Current implementation enforces parts of this via validation and repository/handler logic. This should later be tightened with explicit tests.

---

## 5. Rules for Attempts

- an attempt belongs to exactly one user and one test
- only the owning student may interact with the attempt
- completed attempts must remain reviewable by the owning student
- completed attempts should not revert to in-progress state

---

## 6. Rules for Assignments

- assignments must have a due date
- overdue status is derived from pending assignments whose due date has passed
- assignments are informational/task state, not the authoritative source of test scoring

---

## 7. Rules for Review and Feedback

- publication is blocked until human review approval is recorded
- requesting changes removes the test from approved publication state
- feedback issues are workflow items and must retain reporter, target question, and status history context

---

## 8. Rules for AI-Generated Content

- OCR/text extraction may fail partially; generated content must remain reviewable and editable by an admin
- generated questions are drafts until explicitly saved and then approved/published through normal workflows
- AI output must not be treated as trusted truth without validation

---

## 9. Future Fitness Functions Derived from This Spec

Planned automated checks:
- handlers do not access database pools directly
- public route allowlist remains explicit
- all templates render without panic
- only published tests are student-visible by default
- assignment state transitions remain valid
- attempt ownership is enforced on review/take/results endpoints
- review-gated publication remains enforced
- feedback moderation remains restricted to teacher/admin roles