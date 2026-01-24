You are an expert Go developer. I am building a web application using Go (Golang) on the backend and HTMX on the frontend. The application will be deployed to Oracle Cloud Free Tier on an ARM (Ampere) instance.

**Technology Constraints & Stack:**
1.  **Language:** Go 1.22 or newer.
2.  **Router:** Use "github.com/go-chi/chi/v5" for routing.
3.  **Database:** PostgreSQL. Use "github.com/jackc/pgx/v5" as the driver and for connection pooling.
4.  **Frontend:** Use Go's standard `html/template` package. Use HTMX for dynamic behavior (AJAX without writing JS). Use TailwindCSS (via CDN for now) for styling.
5.  **Architecture:**
    - Use a `cmd/server/main.go` entry point.
    - Put business logic in an `internal/` directory.
    - Put HTML templates in a `views/` directory.
    - Use "Dependency Injection" style: Pass the database pool into your handlers/services struct, do not use global variables.
6.  **Deployment:** The app must run in a Docker container. The Dockerfile must use a multi-stage build to result in a tiny "scratch" or "alpine" image. It must be compatible with ARM64 architecture.

**Your Goal:**
Help me scaffold this application step-by-step. Focus on clean, idiomatic Go code. Always handle errors explicitly.