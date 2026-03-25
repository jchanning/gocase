# GoCaSE New Features - Quick Start Guide

## Getting Started

### 1. Build and Run the Application

```bash
cd c:\Users\johnm\OneDrive\Projects\GoCaSE

# Build the application
go build -o gocase.exe ./cmd/server

# Run with database URL
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/gocase?sslmode=disable"
.\gocase.exe

# Or in Docker
docker-compose up --build
```

## Using the New Features

### Creating and Managing Tests (Teachers)

#### Creating a New Test

1. **Login as a teacher** and go to Teacher Dashboard (`/teacher/dashboard`)
2. Click **"Create Test"** button
3. Fill in test details:
   - **Test Title**: Name of your test
   - **Description**: What the test covers
   - **Subject**: Select from predefined subjects
   - **Exam Standard**: Primary, Secondary, GCSE, or A-Level
   - **Difficulty**: Easy, Medium, or Hard
   - **Time Limit**: Minutes allowed (e.g., 10)
   - **Passing Score**: Percentage needed to pass (0-100)

4. **Add Questions**:
   - Click "Add Question" button for each question
   - Enter question text
   - (Optional) Upload an image for the question
   - Set points for the question
   - Add 4 answer options
   - **Mark one option as correct** by selecting the radio button
   - Remove unwanted questions with "Remove" button

5. Click **"Create Test"** to save

#### Editing a Test

1. Go to **Teacher Dashboard**
2. Find your test in the list
3. Click **"Edit"** button
4. Make changes to test information
5. Click **"Save Changes"**

#### Previewing a Test

Before publishing, check how students will see your test:

1. Click **"Preview"** button on the edit page
2. Review:
   - All test metadata
   - All questions with images
   - All answer options with correct answers highlighted (green)
   - Points for each question
3. Click back to edit if you need to make changes

#### Publishing a Test

Once you're happy with your test:

1. Click **"Publish Test"** button on the edit page
2. Confirm the action
3. Test becomes available to students
4. Status changes to "Published" (green)

#### Unpublishing or Deleting

- **Unpublish**: Click "Unpublish Test" to hide from students (keeps test in draft)
- **Delete**: Click "Delete" to permanently remove (cannot be undone)

### Image Upload for Questions

#### Adding Images to Questions

1. When creating/editing a test, each question has an **"Question Image"** field
2. Click to select an image file
3. **Supported formats**: JPG, JPEG, PNG, GIF, WebP
4. **Max file size**: 32MB per entire form submission
5. Images are stored securely and displayed when students take the test

#### Tips for Images

- Keep images reasonably sized (optimize before upload)
- Use images for geometry, diagrams, charts, etc.
- Images appear above the question text during the test
- Images are displayed at original aspect ratio

### Admin User Management

#### Accessing User Management

1. **Login as an admin**
2. Go to **"/admin/users"** or click "User Management" menu
3. You'll see:
   - List of all users with their roles
   - Form to create new users
   - Actions to manage existing users

#### Creating New Users

1. Fill in the "Create New User" form:
   - **Email**: User's email address (must be unique)
   - **Username**: Display name
   - **Password**: Initial password (user can change after login)
   - **Role**: 
     - Student: Can only take tests
     - Teacher: Can create and manage tests
     - Admin: Full system access

2. Click **"Create User"**
3. New user receives initial credentials

#### Changing User Roles

1. Find the user in the table
2. Click the **Role** dropdown
3. Select new role: Student, Teacher, or Admin
4. Changes apply immediately

#### Resetting Passwords

1. Click **"Reset Password"** for the user
2. Enter new password in the prompt
3. Password is updated and hashed immediately
4. User uses new password on next login

#### Deleting Users

1. Click **"Delete"** for the user
2. Confirm deletion
3. **Warning**: This permanently removes the user and their data
4. Cannot be undone

### Role-Based Account Creation

#### For Students

- Can self-register at the register page
- Only option is Student role
- Immediately can take published tests

#### For Teachers and Admins

- **Cannot** self-register as Teacher or Admin
- **Only admins** can create Teacher and Admin accounts
- Must use the Admin User Management page

#### Example Workflow

```
1. Admin registers themselves (becomes Admin)
2. Admin creates Teacher accounts for staff
3. Admin creates Student accounts for pupils (or students self-register)
4. Teachers log in and create tests
5. Teachers publish tests for students to take
6. Students log in and take available tests
```

## Validation and Error Handling

### Form Validation

The system validates data before saving:

- **Test Title**: Required, max 255 characters
- **Description**: Required
- **Exam Standard**: Must be valid (Primary, Secondary, GCSE, A-Level)
- **Difficulty**: Must be valid (Easy, Medium, Hard)
- **Time Limit**: Must be greater than 0
- **Passing Score**: Must be 0-100%

### Question Validation

- **Question Text**: Required, max 5000 characters
- **Answer Options**: Exactly 4 required
- **Correct Answer**: Exactly one must be marked as correct
- **Points**: Must be greater than 0

### Image Validation

- **File Type**: Must be JPG, JPEG, PNG, GIF, or WebP
- **File Size**: Under 32MB total per form submission
- Invalid files are skipped without error

## File Structure

Test images are stored at:
```
assets/
└── uploads/
    └── {testID}/
        ├── question_1.jpg
        ├── question_2.png
        └── question_3.gif
```

## Database Changes

### Test Table Updates

The `tests` table now includes:
- `published BOOLEAN DEFAULT FALSE` - Whether test is available to students
- Questions and answer options remain the same structure

Existing data is not affected.

## API Endpoints Summary

### Teacher Endpoints

```
GET  /teacher/test/create              - Create test form
POST /teacher/test/create              - Submit new test  
GET  /teacher/test/{id}/edit           - Edit test form
POST /teacher/test/{id}/update         - Update test
GET  /teacher/test/{id}/preview        - Preview test
POST /teacher/test/{id}/publish        - Publish
POST /teacher/test/{id}/unpublish      - Unpublish
POST /teacher/test/{id}/delete         - Delete test
```

### Admin Endpoints

```
GET  /admin/users                      - User management page
POST /admin/users/create               - Create user
POST /admin/users/{id}/role            - Update role
POST /admin/users/{id}/reset-password  - Reset password
POST /admin/users/{id}/delete          - Delete user
```

## Troubleshooting

### Test Creation Not Saving
- Check all required fields are filled
- Verify you're logged in as a teacher or admin
- Check browser console for validation errors

### Image Not Uploading
- Verify file is in supported format (JPG, PNG, GIF, WebP)
- Check file size isn't too large
- Check filesystem permissions on uploads directory

### Can't Create Teacher Account
- Only admins can create teacher accounts
- Use admin panel at `/admin/users`
- Teachers cannot register themselves

### User Can't Login After Creation
- Verify email and password are correct
- Check user's role allows that action
- Try resetting password from admin panel

## Security Notes

1. **Passwords**: Hashed with bcrypt, never stored in plain text
2. **File Uploads**: Validated by type and stored safely
3. **Access Control**: Role-based on all sensitive operations
4. **Ownership**: Teachers can only edit their own tests
5. **Admin Only**: User management restricted to admins

## Performance Tips

- Keep question images under 500KB for faster loading
- Use appropriate image formats (PNG for diagrams, JPG for photos)
- Test with 50+ questions to ensure responsive UI
- Use meaningful exam standards and difficulties for filtering

## Next Steps

1. **Create a test subject structure** - Organize questions by topic
2. **Set up initial teachers** - Via admin panel
3. **Create sample tests** - For testing
4. **Upload to students** - Let them start taking tests
5. **Monitor progress** - View test results and analytics (future feature)
