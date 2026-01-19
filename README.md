# GoCaSE - Go Comprehensive Assessment System

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![CI Status](https://img.shields.io/github/actions/workflow/status/jchanning/gocase/ci.yml?branch=main&style=for-the-badge&label=CI)](https://github.com/jchanning/gocase/actions)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://hub.docker.com)

**A comprehensive online assessment platform for educational institutions**

[Features](#features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Contributing](#contributing)

</div>

---

A comprehensive multiple-choice question test platform for students preparing for GCSE and A-Level exams. Built with Go, Chi router, PostgreSQL, HTMX, and TailwindCSS. Optimized for deployment on Oracle Cloud Free Tier ARM64 instances.

## Features

### For Students
- 📝 **Multiple Choice Tests** - Practice with 4-option questions across various subjects
- 📊 **Progress Tracking** - Monitor your scores, completed tests, and improvement over time
- 🏆 **Gamification** - Earn badges, points, and achievements for completing tests
- ⚡ **Instant Feedback** - Get immediate results (varies by difficulty level)
- ⏱️ **Timed Tests** - Practice under real exam conditions
- 📈 **Performance Analytics** - View detailed statistics and identify areas for improvement

### For Teachers & Admins
- 📤 **Test Upload** - Easily upload tests via JSON format
- 🎯 **Multiple Standards** - Support for Primary, GCSE, and A-Level exams
- 📚 **Subject Management** - Organize tests by subject and topic
- 🎚️ **Difficulty Levels** - Set tests as Easy, Medium, or Hard
- 👥 **User Management** - Manage student, teacher, and admin accounts

### Test Features
- **Standards**: Primary, GCSE, A-Level
- **Subjects**: Mathematics, Science, History, English, Geography (extensible)
- **Difficulty Levels**: Easy (immediate feedback), Medium (immediate feedback), Hard (end-of-test feedback)
- **Question Types**: Text-based with optional images
- **Scoring**: Customizable passing scores and point values
- **Time Limits**: Configurable per test

## Tech Stack

- **Backend**: Go 1.22+
- **Router**: Chi v5
- **Database**: PostgreSQL with pgx/v5 driver
- **Frontend**: HTMX + TailwindCSS
- **Templates**: Go html/template
- **Authentication**: Session-based with bcrypt
- **Deployment**: Docker (multi-stage build for ARM64)

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/                    # Authentication & session management
│   │   ├── middleware.go
│   │   └── session.go
│   ├── database/
│   │   ├── database.go          # Database connection pooling
│   │   └── schema.sql           # Database schema
│   ├── handlers/                # HTTP request handlers
│   │   ├── admin_handler.go
│   │   ├── auth_handler.go
│   │   ├── dashboard_handler.go
│   │   └── test_handler.go
│   ├── models/
│   │   └── models.go            # Domain models
│   ├── repository/              # Data access layer
│   │   ├── attempt_repository.go
│   │   ├── test_repository.go
│   │   └── user_repository.go
│   └── server/
│       └── server.go            # HTTP server and routes
├── views/                       # HTML templates
│   ├── layout.html              # Base template
│   ├── home.html
│   ├── login.html
│   ├── register.html
│   ├── dashboard.html
│   ├── tests_list.html
│   ├── take_test.html
│   ├── test_results.html
│   └── admin.html
├── sample_tests/                # Example test JSON files
├── Dockerfile                   # Multi-stage Docker build
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.22 or newer
- PostgreSQL 14+
- Docker (for containerized deployment)

### Database Setup

1. Create a PostgreSQL database:
```sql
CREATE DATABASE gocase;
```

2. Run the schema:
```bash
psql -U postgres -d gocase -f internal/database/schema.sql
```

### Environment Variables

Set the following environment variable:

```bash
DATABASE_URL=postgres://username:password@localhost:5432/gocase?sslmode=disable
```

### Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Run the application:
```bash
export DATABASE_URL="postgres://username:password@localhost:5432/gocase"
go run cmd/server/main.go
```

3. Open your browser at `http://localhost:8080`

### Default Users

After running the schema, you'll need to create users via the registration page. To create an admin user, you can manually update the role in the database:

```sql
UPDATE users SET role = 'admin' WHERE email = 'your@email.com';
```

## Uploading Tests

### JSON Format

Tests can be uploaded via the Admin dashboard using JSON files. Here's the format:

```json
{
  "title": "Test Title",
  "description": "Test description",
  "subject": "Mathematics",
  "topic": "Algebra",
  "exam_standard": "GCSE",
  "difficulty": "Medium",
  "time_limit_minutes": 15,
  "passing_score": 70,
  "questions": [
    {
      "question_text": "What is 2 + 2?",
      "image_url": "https://example.com/image.png",
      "points": 1,
      "options": ["2", "3", "4", "5"],
      "correct_index": 2
    }
  ]
}
```

**Fields:**
- `exam_standard`: Must be one of: "Primary", "GCSE", "A-Level", "Secondary"
- `difficulty`: Must be one of: "Easy", "Medium", "Hard"
- `correct_index`: 0-based index (0-3) of the correct answer
- `image_url`: Optional URL to an image for the question

### Sample Tests

Sample test files are provided in the `sample_tests/` directory:
- `math_algebra_easy.json` - Basic algebra (GCSE, Easy)
- `science_biology_medium.json` - Cell biology (GCSE, Medium)
- `math_calculus_hard.json` - Calculus concepts (A-Level, Hard)

## Docker Deployment

### Build for ARM64 (Oracle Cloud Ampere)

```bash
docker build -t gocase:latest .
```

### Run the container

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/gocase" \
  gocase:latest
```

### Docker Compose (with PostgreSQL)

```yaml
version: '3.8'
services:
  db:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: gocase
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./internal/database/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:password@db:5432/gocase?sslmode=disable
    depends_on:
      - db

volumes:
  pgdata:
```

## User Roles

- **Student**: Can take tests, view results, track progress
- **Teacher**: Can upload tests + all student permissions
- **Admin**: Full access to all features

## Gamification System

### Achievements
- 🎯 **First Steps** - Complete your first test (10 points)
- ⭐ **Perfect Score** - Score 100% on any test (50 points)
- 🏆 **Test Master** - Complete 10 tests (100 points)
- 📚 **Quick Learner** - Score above 90% on 5 tests (75 points)
- 🔥 **Streak Champion** - Maintain a 5-day study streak (50 points)

### Statistics Tracked
- Total points earned
- Tests completed
- Tests passed
- Current streak
- Best streak
- Average score
- Recent performance trends

## API Endpoints

### Public Routes
- `GET /` - Home page
- `GET /login` - Login page
- `POST /login` - Login submission
- `GET /register` - Registration page
- `POST /register` - Registration submission

### Protected Routes (Student)
- `GET /dashboard` - Student dashboard
- `GET /tests` - List available tests
- `GET /test/start?id=<id>` - Start a test
- `GET /test/take?attempt_id=<id>` - Take test
- `POST /test/answer` - Submit individual answer (AJAX)
- `POST /test/submit` - Submit complete test
- `GET /test/results?attempt_id=<id>` - View results

### Protected Routes (Admin/Teacher)
- `GET /admin` - Admin dashboard
- `POST /admin/upload` - Upload test (JSON)

## Development

The application follows clean architecture principles:

- **Dependency Injection**: Database pool passed to repositories, no global variables
- **Idiomatic Go**: Explicit error handling throughout
- **Repository Pattern**: Clear separation of data access logic
- **Middleware**: Authentication and authorization handled at routing level
- **Template Rendering**: Server-side rendering with HTMX for dynamic updates

## Security Features

- Password hashing with bcrypt
- Session-based authentication
- HttpOnly cookies
- Role-based access control
- SQL injection protection via parameterized queries

## Performance

- Connection pooling for database efficiency
- Static asset caching
- Minimal container size (~15MB)
- Optimized for ARM64 architecture

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
