# GoCaSE Setup Guide

## Quick Start

### 1. Database Setup

Create a PostgreSQL database:
```bash
createdb gocase
```

Or using psql:
```sql
CREATE DATABASE gocase;
```

Run the schema:
```bash
psql -U postgres -d gocase -f internal/database/schema.sql
```

Or on Windows with PowerShell:
```powershell
$env:PGPASSWORD="your_password"
psql -U postgres -d gocase -f internal\database\schema.sql
```

### 2. Environment Configuration

Set the DATABASE_URL environment variable:

**Linux/macOS:**
```bash
export DATABASE_URL="postgres://postgres:password@localhost:5432/gocase?sslmode=disable"
```

**Windows PowerShell:**
```powershell
$env:DATABASE_URL="postgres://postgres:password@localhost:5432/gocase?sslmode=disable"
```

**Windows CMD:**
```cmd
set DATABASE_URL=postgres://postgres:password@localhost:5432/gocase?sslmode=disable
```

### 3. Run the Application

```bash
go run cmd/server/main.go
```

The application will be available at `http://localhost:8080`

## First Time Setup

### Creating Your First User

1. Navigate to `http://localhost:8080`
2. Click "Register"
3. Fill in the registration form
4. You'll be automatically logged in as a student

### Creating an Admin User

To create an admin or teacher account, first register as normal, then update the role in the database:

```sql
-- Make a user an admin
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';

-- Make a user a teacher
UPDATE users SET role = 'teacher' WHERE email = 'teacher@example.com';
```

### Uploading Your First Test

1. Log in as an admin or teacher
2. Click "Admin" in the navigation
3. Use the file upload section
4. Select one of the sample tests from `sample_tests/` directory
5. The test will be uploaded and available for students

## Sample Data

The `sample_tests/` directory contains three example tests:

1. **math_algebra_easy.json** - GCSE Mathematics, Easy difficulty, 5 questions
2. **science_biology_medium.json** - GCSE Science, Medium difficulty, 8 questions
3. **math_calculus_hard.json** - A-Level Mathematics, Hard difficulty, 6 questions

You can upload these tests via the admin interface to get started quickly.

## User Roles Explained

### Student
- Take tests
- View own results
- Track personal progress
- Earn achievements
- View dashboard statistics

### Teacher
- All student permissions
- Upload new tests
- View all tests
- Manage test content

### Admin
- All teacher permissions
- Full system access
- User management (via database)

## Testing the Application

### Manual Testing Checklist

1. ✅ Register a new student account
2. ✅ Login with student credentials
3. ✅ View the dashboard
4. ✅ Browse available tests
5. ✅ Start a test
6. ✅ Answer questions (auto-save feature)
7. ✅ Submit the test
8. ✅ View results
9. ✅ Check achievements earned
10. ✅ Login as admin/teacher
11. ✅ Upload a test via JSON
12. ✅ Verify test appears in student test list

## Troubleshooting

### Database Connection Issues

**Error**: "Failed to initialize database"
- Verify PostgreSQL is running
- Check DATABASE_URL format
- Ensure database exists
- Verify credentials

### Port Already in Use

**Error**: "address already in use"
- Another application is using port 8080
- Change port in `cmd/server/main.go` (line with `:8080`)
- Or stop the other application

### Template Not Found

**Error**: "Error parsing templates"
- Ensure you're running from the project root directory
- Verify `views/` directory exists
- Check file permissions

### Session/Login Issues

**Issue**: Can't stay logged in
- Check browser cookies are enabled
- Verify session middleware is active
- Clear browser cookies and try again

## Development Tips

### Hot Reload

For development, consider using `air` for hot reload:
```bash
go install github.com/cosmtrek/air@latest
air
```

### Database Migrations

When making schema changes:
1. Update `internal/database/schema.sql`
2. Drop and recreate the database
3. Or use a migration tool like `golang-migrate`

### Adding New Features

The codebase is structured for easy extension:
- Add models in `internal/models/`
- Add repositories in `internal/repository/`
- Add handlers in `internal/handlers/`
- Add routes in `internal/server/server.go`
- Add templates in `views/`

## Production Deployment

### Using Docker

See the main README.md for Docker deployment instructions.

### Security Checklist

- [ ] Use strong DATABASE_URL with complex password
- [ ] Enable SSL for database connection
- [ ] Use HTTPS in production
- [ ] Set secure session cookie settings
- [ ] Regularly update dependencies
- [ ] Implement rate limiting
- [ ] Add CSRF protection for forms
- [ ] Regular database backups

### Performance Optimization

- Use connection pooling (already configured)
- Enable database query caching
- Use a reverse proxy (Nginx/Caddy)
- Enable gzip compression
- Optimize images in questions
- Consider Redis for sessions in production

## Getting Help

If you encounter issues:
1. Check the troubleshooting section above
2. Review the logs for error messages
3. Ensure all prerequisites are installed
4. Verify environment variables are set correctly
