package fitness

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerLayerDoesNotImportDatabaseOrPGXPool(t *testing.T) {
	pattern := filepath.Join("..", "handlers", "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob handlers: %v", err)
	}

	blockedImports := map[string]struct{}{
		"github.com/jackc/pgx/v5":                       {},
		"github.com/jackc/pgx/v5/pgxpool":               {},
		"github.com/jchanning/gocase/internal/database": {},
	}

	fs := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if _, blocked := blockedImports[importPath]; blocked {
				t.Fatalf("handler file %s imports blocked package %s", path, importPath)
			}
		}
	}
}

func TestPublicRouteAllowlistRemainsExplicit(t *testing.T) {
	path := filepath.Join("..", "server", "server.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read server.go: %v", err)
	}

	source := string(content)
	expected := []string{
		`s.router.Get("/", s.handleHome)`,
		`s.router.Get("/login", authHandler.ShowLogin)`,
		`s.router.Post("/login", authHandler.Login)`,
		`s.router.Get("/register", authHandler.ShowRegister)`,
		`s.router.Post("/register", authHandler.Register)`,
	}

	for _, snippet := range expected {
		if !strings.Contains(source, snippet) {
			t.Fatalf("expected public route declaration missing: %s", snippet)
		}
	}

	blocked := []string{
		`s.router.Get("/tests",`,
		`s.router.Get("/dashboard",`,
		`s.router.Get("/admin",`,
	}

	for _, snippet := range blocked {
		if strings.Contains(source, snippet) {
			t.Fatalf("protected route appears to be declared at root router level: %s", snippet)
		}
	}
}

func TestCoreArchitectureDocumentsExist(t *testing.T) {
	docs := []string{
		filepath.Join("..", "..", "docs", "ARCHITECTURE.md"),
		filepath.Join("..", "..", "docs", "API.md"),
		filepath.Join("..", "..", "docs", "IMPLEMENTATION_STATUS.md"),
		filepath.Join("..", "..", "docs", "BLUEPRINT.md"),
		filepath.Join("..", "..", "docs", "MASTER_PLAN.md"),
		filepath.Join("..", "..", "docs", "NON_GOALS.md"),
		filepath.Join("..", "..", "docs", "DOMAIN_SPEC.md"),
	}

	for _, path := range docs {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected architecture document missing: %s", path)
		}
	}
}

func TestHandlerFilesDeclareHandlerPackage(t *testing.T) {
	pattern := filepath.Join("..", "handlers", "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob handlers: %v", err)
	}

	fs := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fs, path, nil, parser.PackageClauseOnly)
		if err != nil {
			t.Fatalf("parse package clause for %s: %v", path, err)
		}

		if file.Name == nil || file.Name.Name != "handlers" {
			t.Fatalf("expected %s to declare package handlers", path)
		}
	}
}

func TestServerFileContainsProtectedGroup(t *testing.T) {
	path := filepath.Join("..", "server", "server.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read server.go: %v", err)
	}

	source := string(content)
	if !strings.Contains(source, `r.Use(authMiddleware.RequireAuth)`) {
		t.Fatalf("expected protected route group to require authentication")
	}
}

func TestAuthAndDashboardHandlersUseInterfacesInsteadOfConcreteRepositories(t *testing.T) {
	files := []string{
		filepath.Join("..", "handlers", "auth_handler.go"),
		filepath.Join("..", "handlers", "dashboard_handler.go"),
		filepath.Join("..", "handlers", "test_handler.go"),
		filepath.Join("..", "handlers", "teacher_handler.go"),
		filepath.Join("..", "handlers", "admin_handler.go"),
	}

	blocked := []string{
		"*repository.UserRepository",
		"*repository.TestRepository",
		"*repository.AttemptRepository",
		"*repository.AssignmentRepository",
	}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		source := string(content)
		for _, snippet := range blocked {
			if strings.Contains(source, snippet) {
				t.Fatalf("expected %s to avoid concrete repository dependency %s", path, snippet)
			}
		}
	}
}

// Compile-time anchor to ensure the package imports go/ast intentionally.
var _ ast.Node
