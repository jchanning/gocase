# Feature Specification: Content Review, Student Issue Feedback, and Create & Manage Consolidation

---

## 1. Feature Name

Content Review, Student Issue Feedback, and Create & Manage Consolidation

## 2. Goal

Introduce a governed content-quality workflow so tests cannot be published without an explicit review record,
students can report unclear or incorrect questions after completing a test, and teachers/admins can manage
creation, review, and follow-up work from a single staff-oriented screen.

## 3. Scope

Included in this feature:

1. Add a test review and approval workflow before publication.
2. Record who reviewed a test, what decision was made, and when it happened.
3. Allow students to report a question or explanation issue from results/review flows.
4. Add a teacher/admin feedback queue to review, triage, resolve, and track reported issues.
5. Combine the current create and system areas into one shared Create & Manage screen for teachers and admins.
6. Preserve the existing server-rendered Go + HTMX architecture and repository boundaries.

## 4. Non-Scope

Explicitly out of scope for this implementation:

1. Email notifications, push notifications, or reminder workflows.
2. Real-time collaboration or multi-user moderation.
3. Public API exposure for feedback/review workflows.
4. AI auto-resolution of reported issues.
5. Large-scale information architecture rewrite beyond consolidating staff create/manage screens.
6. Replacing the current auth/session model or monolith structure.

## 5. Affected Users / Roles

- student
- teacher
- admin

## 6. Constraints

- Must follow the existing SSR architecture in [docs/MASTER_PLAN.md](docs/MASTER_PLAN.md).
- Handler -> repository -> database boundaries remain intact; no direct DB access from handlers.
- Student test visibility rules from [docs/DOMAIN_SPEC.md](docs/DOMAIN_SPEC.md) remain true.
- AI may assist content generation, but it must not bypass review or publish directly.
- The workflow must work for both manually created and AI-generated tests.
- Teacher/admin shared screens should be role-aware, but not branch into separate product flows unless authorization requires it.
- Existing routes should be redirected or preserved where practical to avoid breaking staff bookmarks.

## 7. Data Model Impact

### 7.1 Tests review state

Add current-state review metadata to `tests`:

- `review_status` — suggested values: `draft`, `pending_review`, `approved`, `changes_requested`
- `reviewed_by` — nullable FK to `users.id`
- `reviewed_at` — nullable timestamp
- `review_notes` — nullable text for reviewer response/decision notes
- `submitted_for_review_at` — nullable timestamp

Rationale:

- `published` remains the visibility flag.
- `review_status` becomes the governance flag.
- A test may only transition to `published = true` when `review_status = approved`.

### 7.2 Review history audit

Add a new `test_review_events` table:

- `id`
- `test_id`
- `reviewer_id`
- `decision` (`submitted`, `approved`, `changes_requested`, `unapproved` if needed)
- `notes`
- `created_at`

Rationale:

- The `tests` table provides current status for fast reads.
- `test_review_events` preserves the audit trail required by the feature request.

### 7.3 Student-reported issues

Add a new `test_feedback_issues` table:

- `id`
- `test_id`
- `question_id`
- `attempt_id`
- `reported_by`
- `issue_type` (`incorrect_answer`, `unclear_explanation`, `question_text_issue`, `other`)
- `student_comment`
- `status` (`open`, `in_review`, `resolved`, `dismissed`)
- `review_response`
- `reviewed_by`
- `reviewed_at`
- `created_at`
- `updated_at`

Rationale:

- Links the report to the exact test, question, and attempt context.
- Supports queue triage and staff response without adding a second response table.

### 7.4 Model / repository impact

Expected updates:

- `internal/models/models.go`
- `internal/database/schema.sql`
- `internal/repository/test_repository.go`
- new repository methods and likely a dedicated feedback repository

## 8. Interface / Route Impact

### 8.1 Student flows

Add issue-reporting actions from:

- results page
- review page

Suggested routes:

- `POST /test/feedback/report`
- optional `GET /test/feedback/report` only if a dedicated form page is needed

Form fields:

- `attempt_id`
- `test_id`
- `question_id`
- `issue_type`
- `student_comment`

### 8.2 Staff review flows

Add shared staff feedback queue routes under the teacher/admin protected group.

Suggested routes:

- `GET /manage`
- `GET /manage/tests`
- `GET /manage/feedback`
- `POST /manage/test/{id}/submit-review`
- `POST /manage/test/{id}/approve`
- `POST /manage/test/{id}/request-changes`
- `POST /manage/feedback/{id}/status`
- `POST /manage/feedback/{id}/response`

### 8.3 Existing route consolidation

Current screens to consolidate:

- `/teacher/test/create`
- `/teacher/upload`
- `/admin`
- `/admin/manage`
- `/admin/generate`
- `/admin/wizard`

Plan:

1. Introduce a shared `Create & Manage` entry point.
2. Render tabbed or sectional SSR navigation inside the shared screen.
3. Keep old routes temporarily and redirect them to the relevant section/query state.

### 8.4 Template impact

Likely templates to create or revise:

- new shared staff screen template, likely `views/manage.html`
- `views/create_test.html`
- `views/admin.html`
- `views/admin_manage.html`
- `views/admin_wizard.html`
- `views/test_results.html`
- `views/test_review.html`
- `views/layout.html`

## 9. Business Rules / Invariants

Add or enforce the following rules:

1. A test must not be published unless it has an approved review state.
2. A review decision must record reviewer identity and timestamp.
3. Student issue reports may only be filed by the owner of the attempt.
4. Student issue reports must reference a valid test/question pairing from that attempt.
5. Only teachers/admins may review issues or change issue status.
6. AI-generated content remains draft or pending review until a human approves it.
7. Students must not see internal review notes or staff-only feedback resolution controls.
8. Published-test visibility rules for students remain unchanged.

Recommended domain-spec additions:

- `INV-011 — Review gate before publication`
- `INV-012 — Review audit integrity`
- `INV-013 — Feedback ownership`
- `INV-014 — Feedback moderation authority`

## 10. Edge Cases

1. Legacy published tests exist without approval metadata.
2. A student submits duplicate issue reports for the same question.
3. A report is filed for a deleted or edited question after an attempt was completed.
4. A teacher tries to publish without first submitting for review.
5. An admin resolves a report after the underlying question has already been edited.
6. Teachers should only manage their own tests unless elevated admin access applies.
7. Feedback queue empty state must be explicit and usable.
8. Staff response text may be blank when only status changes are required.

## 11. Acceptance Tests

1. Given an unpublished test with `review_status = draft`, when a teacher tries to publish it, then publication is rejected with a clear error.
2. Given a pending-review test, when a teacher or admin approves it, then the test stores reviewer identity and timestamp and can be published.
3. Given an approved test, when it is published, then students can see it but they still cannot see review metadata.
4. Given a completed student attempt, when the student reports a question issue, then the report is stored against the correct attempt, test, and question.
5. Given a student who does not own an attempt, when they try to report an issue for it, then the request is rejected.
6. Given an open feedback issue, when a teacher/admin marks it resolved with a response, then the queue reflects the new status, reviewer, and response.
7. Given a teacher/admin opening Create & Manage, when they navigate between sections, then they can create tests, view existing tests, review pending content, and access feedback from one screen.
8. Given an old staff route such as `/admin/manage`, when it is opened after rollout, then it redirects or renders the matching Create & Manage section.

## 12. TDD Plan

### Phase 1 — Domain + schema guardrails

- Add failing repository tests for review-state persistence and feedback issue persistence.
- Add failing handler tests for publish gating and feedback authorization.
- Add or extend architecture/fitness tests to require review-gated publication behavior.

Likely test files:

- `internal/repository/test_repository_test.go`
- new `internal/repository/feedback_repository_test.go`
- `internal/handlers/staff_handler_test.go`
- `internal/handlers/test_handler_flow_test.go`
- `internal/handlers/test_handler_unit_test.go`
- `internal/fitness/architecture_test.go`

### Phase 2 — Minimum viable review workflow

- Implement schema and repository changes.
- Implement `submit for review`, `approve`, and `request changes` actions.
- Block publish if review state is not approved.

### Phase 3 — Student issue reporting

- Add the student form/action from results/review.
- Persist issues with ownership validation.

### Phase 4 — Staff feedback queue

- Build queue filtering and detail rendering.
- Add status update and response actions.

### Phase 5 — Create & Manage consolidation

- Introduce shared route/template.
- Migrate existing create/system/generate/wizard entry points behind shared navigation.
- Add redirects from legacy routes.

### Phase 6 — Refactor and documentation

- Update architecture/docs/specs.
- Add migration/backfill notes and rollout validation.

## 13. Documentation Updates

Must update as part of implementation:

- `docs/IMPLEMENTATION_STATUS.md`
- `docs/API.md`
- `docs/ARCHITECTURE.md`
- `docs/DOMAIN_SPEC.md`
- `CHANGELOG.md`

## 14. Rollout Notes

### 14.1 Migration / backfill strategy

The repo already contains published tests created before review governance existed. Implementation must choose one of these rollout options explicitly:

1. Backfill legacy published tests as approved with a system note, preserving availability.
2. Mark legacy published tests as approved-but-unattributed, with nullable reviewer fields and an audit note.
3. Force legacy content through re-review before future edits or re-publication.

Recommended approach:

- Keep existing published tests available.
- Backfill them to an approved current state with an explicit legacy audit event so rollout does not hide existing student content.

### 14.2 Authorization decisions to confirm before implementation

1. Can the creator approve their own test, or must reviewer and creator differ?
2. Should teachers gain access to the existing AI generation workflow, or should the combined screen show role-limited sections?
3. Should duplicate student reports be merged or simply listed separately?

### 14.3 Backward compatibility

- Preserve current URLs during the first rollout via redirects.
- Avoid removing old templates until the shared screen is stable.
- Ensure deployment includes schema migration before app startup uses the new fields.

## 15. Delivery Plan

### Epic A — Review-Gated Publication

Deliverables:

1. Schema for review status and review events.
2. Repository methods for submit/approve/request changes.
3. Publish gate in staff handlers.
4. Staff UI indicators for draft, pending review, approved, and changes requested.

### Epic B — Student Issue Reporting

Deliverables:

1. Report issue action on results/review screens.
2. Feedback issue persistence and validation.
3. Confirmation UI for successful submission.

### Epic C — Staff Feedback Operations

Deliverables:

1. Shared feedback queue with filters by status/test/date.
2. Detail panel or page showing test, question, attempt context, and student comments.
3. Status transitions and reviewer response workflow.

### Epic D — Create & Manage Unification

Deliverables:

1. Single staff landing screen with sections for create, manage, review, AI generation, and feedback.
2. Navigation updates in `layout.html`.
3. Redirects from legacy routes.

### Recommended implementation order

1. Epic A
2. Epic B
3. Epic C
4. Epic D

Reasoning:

- Review-gated publication is the strongest domain rule and should exist before UI consolidation.
- Student reporting and staff moderation share the same data contract, so they should be implemented before finalizing the consolidated screen.
- UI consolidation is safest after the new workflows already exist and can be assembled into one staff surface.