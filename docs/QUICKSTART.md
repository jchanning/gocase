# 🚀 Quick Start Guide

Get GoCaSE running in under 5 minutes!

## Option 1: Docker Compose (Recommended - Fastest)

```bash
# Start everything (database + app)
docker-compose up -d

# Wait ~10 seconds for database initialization
# Then open http://localhost:8080

# To stop
docker-compose down
```

That's it! The database is automatically initialized with the schema.

---

## Option 2: Local Development

### Step 1: Setup Database

**Create database:**
```bash
createdb gocase
```

**Initialize schema:**
```bash
# Windows PowerShell
$env:PGPASSWORD="your_password"
psql -U postgres -d gocase -f internal\database\schema.sql

# Linux/Mac
psql -U postgres -d gocase -f internal/database/schema.sql
```

### Step 2: Set Environment Variable

**Windows PowerShell:**
```powershell
$env:DATABASE_URL="postgres://postgres:password@localhost:5432/gocase?sslmode=disable"
```

**Linux/Mac:**
```bash
export DATABASE_URL="postgres://postgres:password@localhost:5432/gocase?sslmode=disable"
```

### Step 3: Run Application

```bash
go run cmd/server/main.go
```

**Or build and run:**
```bash
go build -o gocase.exe ./cmd/server
./gocase.exe
```

Open http://localhost:8080

---

## First Steps

### 1. Register an Account
- Click "Register" on the home page
- Fill in username, email, password
- You'll be logged in automatically as a **Student**

### 2. Make Yourself an Admin
```sql
-- In psql or your database tool
UPDATE users SET role = 'admin' WHERE email = 'your@email.com';
```

### 3. Upload Sample Tests
- Login as admin
- Click "Admin" in navigation
- Upload files from `sample_tests/` folder:
  - `math_algebra_easy.json`
  - `science_biology_medium.json`
  - `math_calculus_hard.json`

### 4. Take a Test
- Click "Tests" in navigation
- Choose a test
- Click "Start Test"
- Answer questions (they auto-save!)
- Click "Submit Test" when done
- View your results and earned achievements

---

## Troubleshooting

### Can't connect to database?
```bash
# Check PostgreSQL is running
pg_ctl status

# Or on Linux
sudo systemctl status postgresql
```

### Port 8080 already in use?
Edit `cmd/server/main.go` and change `:8080` to another port (e.g., `:3000`)

### Docker issues?
```bash
# Check logs
docker-compose logs app

# Restart services
docker-compose restart
```

---

## Test Features to Try

✅ **Timer** - Tests have time limits (watch the countdown!)  
✅ **Auto-save** - Answers save automatically as you click  
✅ **Achievements** - Complete tests to earn badges  
✅ **Difficulty Modes**:
   - Easy/Medium: See correct answers immediately
   - Hard: Answers shown only at the end  
✅ **Dashboard** - Track your progress over time  
✅ **Streaks** - Study daily to build your streak  

---

## Quick Reference

| URL | Purpose |
|-----|---------|
| `/` | Home page |
| `/login` | Login |
| `/register` | Register new account |
| `/dashboard` | Your progress dashboard |
| `/tests` | Browse available tests |
| `/admin` | Upload tests (admin/teacher only) |

---

## Default Credentials (After Manual Setup)

No default credentials - register your own account first, then promote to admin via SQL.

---

## Need More Help?

📖 See [SETUP.md](SETUP.md) for detailed instructions  
📚 See [README.md](README.md) for full documentation  
📝 See [IMPLEMENTATION.md](IMPLEMENTATION.md) for technical details

---

## Sample Test Upload

Try uploading this minimal test via the admin panel:

```json
{
  "title": "Quick Test",
  "description": "A simple test",
  "subject": "General",
  "topic": "Sample",
  "exam_standard": "GCSE",
  "difficulty": "Easy",
  "time_limit_minutes": 5,
  "passing_score": 60,
  "questions": [
    {
      "question_text": "What is 1 + 1?",
      "points": 1,
      "options": ["1", "2", "3", "4"],
      "correct_index": 1
    }
  ]
}
```

Save this as `test.json` and upload it through the admin interface.

---

Happy Testing! 🎓
