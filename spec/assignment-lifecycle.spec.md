# Assignment Lifecycle Specification

## 1. Feature Name

Teacher/Admin Assignment Lifecycle

## 2. Goal

Allow teachers and admins to assign tests to students with due dates, track pending and overdue work, and mark assignments complete when the assigned student completes the relevant test.

## 3. Scope

- assignment form and submission
- due date capture
- pending/overdue/completed state handling
- teacher dashboard visibility
- student dashboard visibility for assignments

## 4. Non-Scope

- group assignment engine beyond explicit selected students
- email, SMS, or push reminders
- parent notifications
- calendar sync

## 5. Affected Users / Roles

- teacher
- admin
- student

## 6. Constraints

- only teacher/admin may create assignments
- assignments must have a due date
- overdue state is derived from due date and incomplete status
- completion should be linked to actual student test completion, not manual toggles alone

## 7. Data Model Impact

- `test_assignments`
- dashboard views derived from assignment status and due dates

## 8. Interface / Route Impact

- `GET /teacher/test/{id}/assign`
- `POST /teacher/test/{id}/assign`
- dashboard surfaces for teacher and student assignment visibility

## 9. Business Rules / Invariants

- assignment target must be a valid student
- assignment creator must be teacher or admin
- pending assignments whose due date has passed become overdue
- completing the assigned test should mark matching pending/overdue assignment completed
- completed assignments remain historically visible but should not be treated as active work

## 10. Edge Cases

- invalid student IDs in form submission
- missing due date
- duplicate assignment attempts for the same test/student
- assignment exists but student completes a different test

## 11. Acceptance Tests

1. Given a teacher and a valid student selection, when the assignment form is submitted with a due date, then assignment records are created.
2. Given no due date, when the assignment form is submitted, then the request fails with `400 Bad Request`.
3. Given a pending assignment whose due date has passed, when assignments are fetched, then its status is updated to overdue.
4. Given a matching pending assignment, when the student completes the assigned test, then the assignment is marked completed.

## 12. TDD Plan

- add handler tests for assignment form validation and redirects
- add repository tests for overdue marking and completion updates
- add dashboard tests to ensure pending/overdue items surface correctly

## 13. Documentation Updates

- `docs/API.md`
- `docs/DOMAIN_SPEC.md`
- `docs/IMPLEMENTATION_STATUS.md`

## 14. Rollout Notes

- reminders remain deferred until a notification architecture exists
- if assignment groups are added, this spec must be extended before implementation
