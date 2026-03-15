# GoCaSE — Route & API Reference

**Version:** 0.1.0
**Last Updated:** 2026-03-14

All routes are server-side rendered (HTML responses) unless noted. There is no REST/JSON API —
HTMX is used for dynamic interactions.

---

## Authentication

| Method | Path | Handler | Auth Required | Notes |
|--------|------|---------|---------------|-------|
| GET | `/` | `server.handleHome` | No | Public landing page |
| GET | `/login` | `AuthHandler.ShowLogin` | No | Login form |
| POST | `/login` | `AuthHandler.Login` | No | Sets session cookie on success |
| GET | `/register` | `AuthHandler.ShowRegister` | No | Registration form |
| POST | `/register` | `AuthHandler.Register` | No | Students only; elevated roles require admin |
| GET | `/logout` | `AuthHandler.Logout` | Yes | Clears session cookie |

---

## Student Routes

| Method | Path | Handler | Role | Notes |
|--------|------|---------|------|-------|
| GET | `/dashboard` | `DashboardHandler.ShowDashboard` | Any auth | Shows assignments, stats, recent activity |
| GET | `/tests` | `TestHandler.ListTests` | Any auth | Browse published tests; supports groupBy, filters |
| GET | `/test/start` | `TestHandler.StartTest` | student | `?id=<test_id>` — creates attempt, redirects to take |
| GET | `/test/take` | `TestHandler.TakeTest` | student | `?attempt_id=<id>` — renders current question |
| POST | `/test/answer` | `TestHandler.SubmitAnswer` | student | HTMX: saves answer, returns next question fragment |
| POST | `/test/submit` | `TestHandler.SubmitTest` | student | Finalises attempt, calculates score |
| GET | `/test/results` | `TestHandler.ViewResults` | student | `?attempt_id=<id>` — shows score + recommendation |
| GET | `/test/review` | `TestHandler.ReviewTest` | student | `?attempt_id=<id>` — shows all Q&A with explanations |
| POST | `/test/feedback/report` | `TestHandler.ReportIssue` | student | Report an issue against a question/explanation from a completed attempt |
| GET | `/history` | `TestHandler.History` | Any auth | Past attempts with filters |
| GET | `/tests/{id}/notes` | `AdminHandler.ServeTestNotes` | Any auth | Serves uploaded notes file (PDF/PPTX) |

### Query Parameters — `/tests`
| Param | Values | Default |
|-------|--------|---------|
| `subject_id` | integer | — |
| `difficulty` | Easy, Medium, Hard | — |
| `standard` | GCSE, A-Level, Primary, Secondary, IGCSE | — |
| `published` | true, false | `true` |

---

## Staff Routes

> Requires role: `teacher` or `admin`

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/manage` | `ManageHandler.ShowManage` | Combined Create & Manage screen for staff |
| GET | `/admin` | `ManageHandler.ShowManage` | Legacy create route now mapped to Create & Manage |
| GET | `/admin/manage` | `ManageHandler.ShowManage` | Legacy system route now mapped to Create & Manage |
| POST | `/manage/test/{id}/submit-review` | `ManageHandler.SubmitForReview` | Submit test into review queue |
| POST | `/manage/test/{id}/approve` | `ManageHandler.ApproveTest` | Approve test and record reviewer audit fields |
| POST | `/manage/test/{id}/request-changes` | `ManageHandler.RequestChanges` | Move test out of approval state with reviewer notes |
| POST | `/manage/feedback/{id}/update` | `ManageHandler.UpdateFeedbackIssue` | Review, respond to, or resolve student-reported issues |

---

## Teacher Routes

> Requires role: `teacher` or `admin`

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/teacher/dashboard` | `TeacherHandler.ShowDashboard` | Class overview + assignments |
| GET | `/teacher/upload` | `TeacherHandler.ShowUpload` | Upload test via JSON form |
| POST | `/teacher/upload` | `TeacherHandler.UploadTest` | Accepts JSON test file |
| GET | `/teacher/test/create` | `TeacherHandler.ShowCreateTest` | Manual test creation form |
| POST | `/teacher/test/create` | `TeacherHandler.CreateTest` | Creates test + questions |
| GET | `/teacher/test/{id}/edit` | `TeacherHandler.EditTest` | Edit existing test |
| POST | `/teacher/test/{id}/update` | `TeacherHandler.UpdateTest` | Save test edits |
| GET | `/teacher/test/{id}/preview` | `TeacherHandler.PreviewTest` | Preview test as student would see it |
| POST | `/teacher/test/{id}/publish` | `TeacherHandler.PublishTest` | Make test visible to students |
| POST | `/teacher/test/{id}/unpublish` | `TeacherHandler.UnpublishTest` | Hide test from students |
| POST | `/teacher/test/{id}/delete` | `TeacherHandler.DeleteTest` | Permanently delete test |
| DELETE | `/teacher/test/{id}` | `TeacherHandler.DeleteTest` | HTMX delete variant |
| GET | `/teacher/test/{id}/assign` | `TeacherHandler.ShowAssignTest` | Assignment form (select students, due date) |
| POST | `/teacher/test/{id}/assign` | `TeacherHandler.AssignTest` | Create assignment records |

---

## Admin Routes

> Requires role: `admin` (unless shared with teacher, noted above)

### Test Management
| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| POST | `/admin/manage/subjects` | `AdminHandler.CreateSubject` | Create new subject |
| DELETE | `/admin/manage/subjects/{id}` | `AdminHandler.DeleteSubject` | Delete subject (HTMX) |
| GET | `/admin/test/{id}/edit` | `AdminHandler.EditTest` | Edit test (admin view) |
| GET | `/admin/test/{id}/preview` | `TeacherHandler.PreviewTest` | Shared preview handler |
| GET | `/admin/test/{id}/pdf` | `AdminHandler.ExportTestPDF` | Export test as PDF |
| POST | `/admin/test/{id}/update` | `AdminHandler.UpdateTest` | Save test edits |
| POST | `/admin/test/{id}/delete` | `AdminHandler.DeleteTest` | Delete test |
| DELETE | `/admin/test/{id}` | `AdminHandler.DeleteTest` | HTMX delete variant |
| POST | `/admin/test/{id}/remove-notes` | `AdminHandler.RemoveTestNotes` | Remove associated notes file |

### LLM Generation
| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/admin/wizard` | `AdminHandler.ShowWizard` | Test creation wizard UI |
| POST | `/admin/wizard` | `AdminHandler.CreateWizardTest` | Guided test creation |
| GET | `/admin/generate` | `AdminHandler.ShowGenerate` | Notes → MCQ generation form, available from the shared staff surface |
| POST | `/admin/generate` | `AdminHandler.GenerateFromNotes` | Calls OCI GenAI, returns JSON |

### User Management
| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/admin/users` | `AdminHandler.ShowUserManagement` | List all users |
| POST | `/admin/users/create` | `AdminHandler.CreateUser` | Create user with specified role |
| POST | `/admin/users/{id}/role` | `AdminHandler.UpdateUserRole` | Change user role |
| POST | `/admin/users/{id}/reset-password` | `AdminHandler.ResetUserPassword` | Set new password |
| POST | `/admin/users/{id}/delete` | `AdminHandler.DeleteUser` | Delete user account |
| DELETE | `/admin/users/{id}` | `AdminHandler.DeleteUser` | HTMX delete variant |

---

## Static Assets

| Path | Description |
|------|-------------|
| `/assets/*` | CSS, JS, images — served from `./assets/` directory |

---

## Session Cookie

| Property | Value |
|----------|-------|
| Name | `gocase_session` |
| Storage | In-memory `sync.Map` on server |
| Lifetime | 24 hours |
| HttpOnly | Yes |
| SameSite | Lax |

---

## JSON Upload Format (Tests)

Tests can be uploaded via `/admin/upload` or `/teacher/upload` in this format:

```json
{
  "title": "GCSE Maths — Algebra",
  "description": "Practice questions covering linear equations",
  "subject": "Mathematics",
  "exam_standard": "GCSE",
  "difficulty": "Medium",
  "time_limit_minutes": 30,
  "passing_score": 60,
  "questions": [
    {
      "question_text": "Solve for x: 2x + 4 = 10",
      "points": 1,
      "explanation": "Subtract 4 from both sides: 2x = 6, then divide by 2: x = 3",
      "options": [
        { "text": "x = 2", "is_correct": false },
        { "text": "x = 3", "is_correct": true },
        { "text": "x = 4", "is_correct": false },
        { "text": "x = 5", "is_correct": false }
      ]
    }
  ]
}
```

**Constraints:**
- `exam_standard`: one of `Primary`, `Secondary`, `GCSE`, `IGCSE`, `A-Level`
- `difficulty`: one of `Easy`, `Medium`, `Hard`
- Each question must have exactly **4 options** with exactly **1** marked `is_correct: true`
- `passing_score`: integer 0–100 (percentage)
