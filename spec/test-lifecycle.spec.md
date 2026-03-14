# Test Lifecycle Specification

## 1. Feature Name
Test Creation, Publication, Attempt, Results, and Review Lifecycle

## 2. Goal
Define the deterministic lifecycle for tests from authoring through publication, student attempts, scoring, results, and review.

## 3. Scope
- teacher/admin test creation and editing
- publish/unpublish transitions
- student discovery of available tests
- attempt creation and submission
- deterministic scoring and pass/fail logic
- review and recommendation flow

## 4. Non-Scope
- essay/free-text questions
- adaptive testing engine
- AI-controlled scoring
- collaborative authoring

## 5. Affected Users / Roles
- student
- teacher
- admin

## 6. Constraints
- tests are persisted in PostgreSQL
- only published tests are student-visible by default
- questions have exactly four options and one correct answer
- scoring is computed from persisted answer truth only
- AI may help generate question drafts, but not scoring or publication state

## 7. Data Model Impact
- `tests`
- `questions`
- `answer_options`
- `test_attempts`
- `student_answers`

## 8. Interface / Route Impact
- `GET /tests`
- `GET /test/start`
- `GET /test/take`
- `POST /test/answer`
- `POST /test/submit`
- `GET /test/results`
- `GET /test/review`
- teacher/admin edit, publish, unpublish, delete, and preview routes

## 9. Business Rules / Invariants
- a test must have valid metadata before it can be published
- a published test may be taken by students
- a student may only interact with their own attempts
- a completed attempt must not revert to in-progress
- pass/fail is derived from score percentage versus test passing score
- recommendations may suggest a next test but must not mutate persisted state

## 10. Edge Cases
- student tries to start an unpublished test
- attempt ID does not belong to current user
- auto-submit when timer expires
- test has no subject and therefore no recommendation candidate
- explicit published=false filter used by teacher/admin

## 11. Acceptance Tests
1. Given a published test, when a student starts it, then a new `in_progress` attempt is created.
2. Given an attempt owned by another user, when a student requests take/results/review routes, then access is denied.
3. Given submitted answers, when a test is submitted, then score and pass/fail are computed deterministically.
4. Given a passing result and an eligible harder test in the same subject, when results are viewed, then a recommendation is displayed.
5. Given a teacher or admin viewing `/tests` without an explicit published filter, when tests are listed, then both published and unpublished tests are visible.

## 12. TDD Plan
- extend `internal/handlers/test_handler_filters_test.go`
- add handler tests around ownership, submission, and result visibility
- add repository tests for recommendation and publication queries

## 13. Documentation Updates
- `docs/API.md`
- `docs/DOMAIN_SPEC.md`
- `docs/IMPLEMENTATION_STATUS.md`

## 14. Rollout Notes
- if publication rules tighten, add explicit validation and a fitness test
- if question types expand, update this spec before code changes