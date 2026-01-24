package views

import (
	"bytes"
	"html/template"
	"path/filepath"
	"testing"
	"time"

	"my-app/internal/models"
)

// Ensures admin user management template renders successfully with sample data.
func TestAdminUsersTemplateRenders(t *testing.T) {
	basePath := filepath.Join("..", "..", "views")
	layout := filepath.Join(basePath, "layout.html")
	admin := filepath.Join(basePath, "admin_users.html")

	tmpl, err := template.ParseFiles(layout, admin)
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Session": nil,
		"Users": []models.User{
			{
				ID:        1,
				Email:     "admin@example.com",
				Username:  "admin",
				Role:      "admin",
				CreatedAt: time.Now(),
			},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("User Management")) {
		t.Fatalf("rendered template missing expected content")
	}
}
