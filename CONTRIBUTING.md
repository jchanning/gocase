# Contributing to GoCaSE

Thank you for your interest in contributing to GoCaSE! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/jchanning/gocase/issues)
2. If not, create a new issue using the Bug Report template
3. Provide detailed information including steps to reproduce, expected vs actual behavior, and environment details

### Suggesting Features

1. Check if the feature has already been suggested in [Issues](https://github.com/jchanning/gocase/issues)
2. Create a new issue using the Feature Request template
3. Clearly describe the problem you're trying to solve and your proposed solution

### Submitting Code Changes

1. **Fork the repository** and create your branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the code style guidelines below

3. **Test your changes**
   ```bash
   go test ./...
   go vet ./...
   ```

4. **Commit your changes** with clear, descriptive messages
   ```bash
   git commit -m "Add feature: brief description"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** using the PR template

## Development Setup

### Prerequisites

- Go 1.25 or later
- PostgreSQL 14 or later
- Docker (optional, for containerized development)

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/jchanning/gocase.git
   cd gocase
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up the database:
   ```bash
   createdb gocase
   psql gocase < internal/database/schema.sql
   ```

4. Set environment variables:
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gocase?sslmode=disable"
   export SESSION_SECRET="your-secret-key"
   ```

5. Run the application:
   ```bash
   go run ./cmd/server
   ```

### Using Docker

```bash
docker-compose up --build
```

## Code Style Guidelines

### Go Code

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to catch common errors
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and concise

### Naming Conventions

- **Files**: lowercase with underscores (e.g., `user_handler.go`)
- **Packages**: lowercase, single word when possible
- **Variables**: camelCase (e.g., `userName`)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE
- **Functions**: PascalCase for exported, camelCase for unexported

### Project Structure

```
GoCaSE/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── auth/           # Authentication & sessions
│   ├── database/       # Database connection & schema
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Data models
│   └── repository/     # Data access layer
├── views/              # HTML templates
├── assets/             # Static assets (CSS, JS, images)
└── sample_tests/       # Example test files
```

## Testing

### Writing Tests

- Write unit tests for all new functions
- Use table-driven tests when appropriate
- Mock external dependencies
- Aim for >80% code coverage

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

## Database Migrations

When making database changes:

1. Update `internal/database/schema.sql`
2. Document the migration in your PR
3. Ensure backward compatibility when possible

## Commit Message Guidelines

Follow the conventional commits specification:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Example:
```
feat: add email notification for test completion

- Implement email service using SMTP
- Add template for completion notification
- Update user settings to include email preferences
```

## Pull Request Process

1. Update documentation for any changed functionality
2. Add tests for new features or bug fixes
3. Ensure all tests pass and code passes linting
4. Update the README.md if needed
5. Reference related issues in your PR description
6. Request review from maintainers

## Questions?

If you have questions, feel free to:
- Open an issue for discussion
- Reach out to the maintainers

Thank you for contributing to GoCaSE! 🎉
