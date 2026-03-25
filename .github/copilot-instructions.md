# GitHub Copilot Instructions — GoCaSE

These instructions apply to all code generation and review in this repository.

---

## Go Source Files (`**/*.go`)

### Package Structure
- `internal/handlers/` — HTTP handlers only (parse request, call repo, render template)
- `internal/repository/` — all database queries; no business logic
- `internal/models/` — plain data structs; no methods with side effects
- `internal/auth/` — session management and middleware only
- `internal/validation/` — input validation helpers
- `internal/llm/` — OCI GenAI client and prompt construction

### Style
- Use early returns to reduce nesting; prefer `if err != nil { return }` over nested ifs
- Keep functions under ~50 lines; split larger functions with a clear single responsibility
- Exported functions need a doc comment; unexported functions only if non-obvious
- Use `context.Context` as the first argument in all repository methods
- Name errors `err`, database row sets `rows`, loop variables with their type prefix (e.g. `test`, `user`)

### Database / Repository
- All queries must use parameterized arguments — **NEVER** string-interpolate SQL
- Use `pgx/v5` scan patterns with explicit column lists — no `SELECT *`
- Repository methods return `(value, error)` — never log inside a repository
- Close rows with `defer rows.Close()` immediately after the error check

### HTTP Handlers
- Read form values with `r.FormValue("field")` for POST, `r.URL.Query().Get("field")` for GET
- Always check the session via `auth.GetSessionData(r)` before accessing user-specific data
- Render templates by calling `template.ParseFiles` with `"views/layout.html"` and the page template, then execute via `tmpl.ExecuteTemplate(w, "layout.html", data)`
- Return HTTP errors with `http.Error(w, message, statusCode)` — never panic

### Error Handling
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Log errors at the handler boundary only, not in the repository or service layer
- Database "not found" should surface as the `pgx.ErrNoRows` error — callers decide how to handle it

### Security
- Hash passwords with `bcrypt` — minimum cost 12
- Session tokens must be cryptographically random (use `crypto/rand`)
- All user-supplied filenames must be sanitised before filesystem use
- Role checks must use `authMiddleware.RequireRole("admin")` middleware — never perform inline string role comparisons in handlers

---

## HTML Templates (`views/**/*.html`)

### Base Layout
- All page templates must define a `"content"` block consumed by the shared layout:
  ```html
  {{define "content"}}
  ...
  {{end}}
  ```
- The layout is `views/layout.html` and provides: navbar, footer, CSS variables, TailwindCSS CDN

### Template Data
- Template data is `map[string]interface{}` — access with `.KeyName`
- The session object is always passed as `.Session` (may be nil on public pages — always guard with `{{if .Session}}`)
- Available session fields: `.Session.UserID`, `.Session.Username`, `.Session.Role`, `.Session.AppVersion`

### CSS / Styling
- Use Tailwind utility classes — do not write custom inline CSS
- Use the semantic design tokens defined in `layout.html`: `--edu-primary`, `--edu-accent`, `--edu-bg`, `--edu-surface`
- Reusable component classes: `app-card`, `app-input`, `btn`, `btn-primary`, `btn-secondary`, `btn-accent`
- Page headings use `page-title` class; subtitles use `page-subtitle` class

### Security
- Never output raw HTML with `{{.Value | html}}` unless the source is trusted admin-controlled content
- User-supplied content must use `{{.Value}}` (auto-escaped by `html/template`)
- All state-changing forms must POST to a handler that validates session ownership

### Accessibility
- Every `<input>` must have a corresponding `<label for="...">` or `aria-label`
- Buttons must have visible text or an `aria-label`
- Use semantic HTML: `<nav>`, `<main>`, `<footer>`, `<section>`, `<article>` appropriately

---

## Tests (`**/*_test.go`)

### Test Structure
- Use `testing.T` only — no testify/assert (not a project dependency)
- Use table-driven tests (`tests := []struct{...}{}`) for multiple input/output cases
- Helper functions that call `t.Fatal` must accept `t *testing.T` as their first argument
- Use `t.Run("description", func(t *testing.T) {...})` for sub-tests

### What to Test
- **Handler tests**: test the HTTP handler function directly using `httptest.NewRequest` and `httptest.NewRecorder`
- **Filter/parse tests**: pure functions — test all branches including edge cases
- **Repository tests**: use `t.Skip()` with the `-short` flag when a real DB is not available
- **Template tests**: verify templates render without panic by executing against a buffer

### What NOT to Test
- Do not test Go standard library behaviour
- Do not write tests that only assert `err == nil` without verifying the meaningful result
- Do not write tests that require network access without a skip guard

### Mocking
- Use interface-based fakes for repository dependencies in handler tests
- Define fake structs in `test_creation_helper.go` or `*_test.go` files, never in production code
- Fakes should implement only the minimal interface needed, not a full mock framework

### Integration Tests
- Tests that require a real PostgreSQL database must be guarded with:
  ```go
  if testing.Short() {
      t.Skip("skipping integration test in short mode")
  }
  ```
- The integration test helper in `internal/handlers/test_creation_helper.go` provides `setupTestDB()`

### Test File Conventions
- Unit tests: same package as the code under test (e.g. `package handlers`)
- Integration tests: same file, guarded with `testing.Short()`
- Use white-box (same-package) style to retain access to unexported internals — avoid `package handlers_test`
