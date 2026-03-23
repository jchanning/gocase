# Complete Change Log - All Implementations

## v0.1.3 — 2026-03-23 — Syllabus & Revision Planner + Schema Auto-Migration

### New Features

**Syllabus Management (`/admin/syllabus`)**
- Admin and teacher users can create, edit, publish, and unpublish exam syllabi
- Syllabi are organised into sections, each containing ordered curriculum topics
- Topics carry title, description, estimated hours, and free-text notes
- Topics can be linked to existing tests for direct access from the revision planner
- Test search panel filters by subject when linking tests to topics

**Revision Planner (student-facing)**
- Students can create a personal revision plan for any published syllabus
- Plan inputs: exam date, daily study hours, active study days of the week
- Scheduler auto-generates per-topic sessions proportional to estimated hours
- Session status tracking: scheduled → completed / skipped
- Dashboard shows upcoming sessions and overall plan progress

**Database Routing & Views**
- New routes: `GET/POST /admin/syllabus`, `GET/POST /admin/syllabus/{id}`, publish/unpublish, section CRUD, topic CRUD, test link/unlink
- New views: `syllabus_manage.html`, `syllabus_edit.html`, `revision_planner.html`, `revision_plan.html`, `revision_create.html`

### Bug Fixes

**Schema auto-migration on startup** (`internal/database/database.go`, `cmd/server/main.go`)
- Added `ApplySchema` method that embeds `schema.sql` at compile time and executes it against the database on every startup
- All DDL uses `CREATE TABLE IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS` so it is safe to run against an already-populated database
- Fixes 500 error on `/admin/syllabus` when the database was initialised before the syllabus tables were added to the schema

---

## 2026-02-17 - UI/UX Modernization Rollout

### Visual Design System

**File: `views/layout.html`**

Implemented a semantic education-focused theme system and reusable UI primitives:

- Added semantic color tokens (`--edu-primary`, `--edu-accent`, `--edu-bg`, `--edu-surface`, text/border/status colors)
- Standardized typography and heading utilities (`page-title`, `page-subtitle`)
- Added reusable components (`app-card`, `app-input`, `btn`, `btn-primary`, `btn-secondary`, `btn-accent`)
- Improved focus visibility and keyboard accessibility styles
- Harmonized shadows, radii, and default utility-class color overrides for consistent rendering

### Student & Shared UX Updates

Updated key student-facing and shared journeys to use the new component system:

- `views/home.html`
- `views/login.html`
- `views/register.html`
- `views/tests_list.html`
- `views/dashboard.html`
- `views/take_test.html`
- `views/test_results.html`
- `views/test_review.html`
- `views/history.html`

Highlights:

- Consistent cards, buttons, and form fields across all primary user flows
- Improved filter form labeling and submit semantics
- Improved readability and hierarchy of headings/metadata
- Preserved existing application behavior while modernizing visual treatment

### Teacher & Admin UX Updates

Updated teacher/admin workflows for consistent styling and better responsive behavior:

- `views/teacher_dashboard.html`
- `views/teacher_upload.html`
- `views/admin.html`
- `views/admin_manage.html`
- `views/admin_users.html`
- `views/admin_wizard.html`
- `views/edit_test.html`
- `views/test_preview.html`

Highlights:

- Responsive action bars and form grids for smaller screens
- Standardized dashboard/admin card styling
- Improved wizard and management page form consistency
- Added/adjusted accessible labels and control semantics where diagnostics flagged issues
- Guarded optional admin-page JavaScript initialization to avoid null-element runtime errors

### Quality & Verification

- Template diagnostics for updated views: no active errors
- View template test passing: `internal/views/template_test.go`
- Docker rebuild/recreate completed to ensure updated templates are served (`docker compose up -d --build --force-recreate`)
- Manual QA pass completed by user and confirmed visually good

### Rollout Status

This phase completes:

- Responsive refinements
- Usability and visual QA
- Changelog rollout documentation

## Database Schema Changes

### File: `internal/database/schema.sql`
**Change**: Added `published` column to tests table
```sql
ALTER TABLE tests ADD COLUMN published BOOLEAN DEFAULT FALSE;
```
- Allows tests to be in draft mode before publication
- Students only see published tests

---

## Model Changes

### File: `internal/models/models.go`
**Change**: Added `Published` field to Test struct
```go
type Test struct {
    // ... existing fields ...
    Published bool `json:"published"`
}
```

---

## Handler Changes

### File: `internal/handlers/teacher_handler.go`

**New Methods Added**:
1. `ShowCreateTest()` - Display test creation form
2. `CreateTest()` - Handle test creation form submission
3. `EditTest()` - Display test editing form
4. `UpdateTest()` - Handle test updates
5. `PreviewTest()` - Show test preview to teacher
6. `PublishTest()` - Publish test for students
7. `UnpublishTest()` - Unpublish test
8. `DeleteTest()` - Delete test permanently
9. `saveUploadedImage()` - Handle image uploads

**New Imports**:
```go
import (
    "io"
    "os"
    "path/filepath"
    "strconv"
    "strings"
   "github.com/jchanning/gocase/internal/validation"
)
```

**Features**:
- Multi-part form handling for image uploads
- Image validation (JPG, PNG, GIF, WebP)
- Safe file path handling
- Directory creation with proper permissions
- Question management with dynamic form
- Ownership verification for test edits
- Test publication workflow

---

### File: `internal/handlers/admin_handler.go`

**New Methods Added**:
1. `ShowUserManagement()` - Display user management page
2. `CreateUser()` - Create new user account
3. `UpdateUserRole()` - Change user's role
4. `ResetUserPassword()` - Reset user password
5. `DeleteUser()` - Delete user account

**New Imports**:
```go
import (
    "strconv"
    "golang.org/x/crypto/bcrypt"
)
```

**Features**:
- User CRUD operations
- Role validation
- Password hashing with bcrypt
- User stats initialization
- Admin-only access control

---

### File: `internal/handlers/auth_handler.go`

**No changes** - Already implements role-based registration!

**Existing Features**:
- Restricts teacher/admin account creation to admins only
- Enforces role validation
- Returns 403 Forbidden for unauthorized role creation

---

## Repository Changes

### File: `internal/repository/test_repository.go`

**Updated Methods** (to include `published` field):
1. `GetAll()` - Updated SELECT to include published
2. `GetByID()` - Updated SELECT to include published
3. `GetByCreator()` - Updated SELECT to include published

**New Methods Added**:
1. `Update()` - Update test metadata
2. `PublishTest()` - Set published=true
3. `UnpublishTest()` - Set published=false
4. `DeleteTest()` - Delete test
5. `DeleteQuestion()` - Delete question
6. `UpdateQuestion()` - Update question
7. `UpdateAnswerOption()` - Update answer option
8. `DeleteAnswerOption()` - Delete option

**SQL Changes**:
- All SELECT queries updated to include `published`
- New UPDATE/DELETE operations for test management

---

### File: `internal/repository/user_repository.go`

**New Methods Added**:
1. `GetAllUsers()` - Fetch all users
2. `GetUsersByRole()` - Fetch users filtered by role
3. `UpdateUserRole()` - Change user's role
4. `UpdatePasswordHash()` - Update password
5. `DeleteUser()` - Delete user account

**SQL Operations**:
- SELECT all users with order by creation date
- UPDATE user role
- UPDATE password hash
- DELETE user record

---

## New Files Created

### File: `internal/validation/test_validator.go`

**Purpose**: Comprehensive validation for test operations

**Key Classes**:
- `TestValidator` - Main validator struct
- `ValidationError` - Error representation

**Methods**:
1. `NewTestValidator()` - Constructor
2. `ValidateTest()` - Validates test metadata
3. `ValidateQuestion()` - Validates question structure
4. `ValidateAnswerOption()` - Validates answer options
5. `GetErrors()` - Get all validation errors
6. `GetErrorMessages()` - Get errors as map
7. Helper functions for validation rules

**Validation Rules**:
- Test title: required, max 255 chars
- Description: required
- Exam standard: must be valid
- Difficulty: must be Easy/Medium/Hard
- Time limit: > 0 minutes
- Passing score: 0-100%
- Questions: exactly 4 options with 1 correct
- Options: required text, max 1000 chars
- Points: > 0

---

### File: `views/create_test.html`

**Purpose**: Form for creating new tests

**Key Features**:
- Test metadata form (title, description, subject, etc.)
- Dynamic question addition/removal with JavaScript
- Image upload fields for questions
- Answer option management (4 per question)
- Radio buttons to select correct answer
- Form validation on client side
- Submit and cancel buttons

**JavaScript Functions**:
- `addQuestion()` - Add new question form
- `removeQuestion()` - Remove question form
- Dynamic template generation

---

### File: `views/edit_test.html`

**Purpose**: Edit existing tests

**Key Features**:
- Update test metadata
- Display all questions in read-only format
- Show images with questions
- Highlight correct answers (green)
- Publish/Unpublish buttons
- Preview button
- Delete button with confirmation
- Save changes button

**JavaScript Functions**:
- `publishTest()` - AJAX publish operation
- `unpublishTest()` - AJAX unpublish operation
- `deleteTest()` - AJAX delete with confirmation

---

### File: `views/test_preview.html`

**Purpose**: Preview test exactly as students will see it

**Key Features**:
- Test metadata display
- Publication status indicator
- All questions with images
- All answer options
- Correct answers highlighted (only visible to teachers)
- Points per question
- Professional layout

**Back Link**: Link back to edit page

---

### File: `views/admin_users.html`

**Purpose**: Manage users (create, update, delete, reset passwords)

**Key Features**:
- Create user form (email, username, password, role)
- Users table with all details
- Inline role selector dropdown
- Reset password button
- Delete button with confirmation
- Form validation
- JavaScript AJAX operations

**JavaScript Functions**:
- `updateRole()` - Change user role
- `resetPassword()` - Reset password
- `deleteUser()` - Delete user
- Form submit handler for creating users

---

## Server Routes

### File: `internal/server/server.go`

**New Routes Added**:

**Teacher Routes** (POST /teacher/... routes):
```
GET  /teacher/test/create           ShowCreateTest
POST /teacher/test/create           CreateTest
GET  /teacher/test/{id}/edit        EditTest
POST /teacher/test/{id}/update      UpdateTest
GET  /teacher/test/{id}/preview     PreviewTest
POST /teacher/test/{id}/publish     PublishTest
POST /teacher/test/{id}/unpublish   UnpublishTest
POST /teacher/test/{id}/delete      DeleteTest
DELETE /teacher/test/{id}           DeleteTest (alt)
```

**Admin Routes**:
```
GET  /admin/users                   ShowUserManagement
POST /admin/users/create            CreateUser
POST /admin/users/{id}/role         UpdateUserRole
POST /admin/users/{id}/reset-password ResetUserPassword
POST /admin/users/{id}/delete       DeleteUser
DELETE /admin/users/{id}            DeleteUser (alt)
```

**Route Protection**:
- All routes require authentication
- Teacher routes: require teacher or admin role
- Admin routes: require admin role only

---

## Summary Statistics

### Lines of Code Added
- Handler methods: ~700 lines
- Validation logic: ~200 lines
- HTML views: ~500 lines
- Repository methods: ~300 lines
- **Total: ~1,700 lines**

### Files Modified: 7
- models/models.go
- database/schema.sql
- handlers/teacher_handler.go
- handlers/admin_handler.go
- repository/test_repository.go
- repository/user_repository.go
- server/server.go

### Files Created: 5
- validation/test_validator.go
- views/create_test.html
- views/edit_test.html
- views/test_preview.html
- views/admin_users.html

### Documentation Files: 3
- FEATURES_IMPLEMENTED.md
- QUICK_START_FEATURES.md
- IMPLEMENTATION_COMPLETE.md

---

## Build Status

**Go Build Command**:
```bash
go build -o gocase.exe ./cmd/server
```

**Result**: ✅ **SUCCESS** - No errors, no warnings

**Application Status**: Ready for testing and deployment

---

## Feature Checklist

- ✅ Test creation UI
- ✅ Test editing UI
- ✅ Test data validation
- ✅ Subject/difficulty/standards management
- ✅ Image upload support
- ✅ Test preview feature
- ✅ Admin user management
- ✅ Role-based registration
- ✅ Publish/unpublish workflow
- ✅ Password hashing with bcrypt
- ✅ Role-based access control
- ✅ Input validation
- ✅ Error handling

---

## Testing Recommendations

1. **Test Creation**
   - Create test with various question counts
   - Upload images (try different formats)
   - Verify validation catches errors

2. **Test Editing**
   - Edit test metadata
   - Modify questions
   - Verify changes save

3. **Preview & Publish**
   - Preview test before publishing
   - Publish/unpublish test
   - Verify students only see published tests

4. **User Management**
   - Create users of each role
   - Change roles
   - Reset password
   - Delete user

5. **Security**
   - Verify teachers can't edit others' tests
   - Verify non-admins can't create teacher accounts
   - Verify students can't access admin functions

---

## Performance Considerations

- Image uploads validated before processing
- Safe file path handling prevents directory traversal
- Database queries use proper indexing
- Validation prevents invalid data from database operations
- Role checks prevent unauthorized access

---

## Backward Compatibility

- No breaking changes to existing data
- Existing tests work with new published field (defaults to FALSE)
- Existing users work with new role system
- All queries backward compatible

---

## Future Enhancement Opportunities

1. Test duplication
2. Test scheduling
3. Bulk user import
4. Analytics dashboard
5. Advanced search/filtering
6. Test revision history
7. Batch operations
8. API documentation
9. Integration tests
10. Performance monitoring
