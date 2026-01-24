package server

import (
	"html/template"
	"log"
	"net/http"

	"my-app/internal/auth"
	"my-app/internal/database"
	"my-app/internal/handlers"
	"my-app/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server holds the HTTP server dependencies.
type Server struct {
	db           *database.Service
	router       *chi.Mux
	sessionStore *auth.SessionStore
}

// NewServer creates and configures a new HTTP server.
func NewServer(db *database.Service) *Server {
	s := &Server{
		db:           db,
		router:       chi.NewRouter(),
		sessionStore: auth.NewSessionStore(),
	}

	// Add middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Serve static files from ./assets directory
	s.router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Pool())
	testRepo := repository.NewTestRepository(db.Pool())
	attemptRepo := repository.NewAttemptRepository(db.Pool())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, s.sessionStore)
	dashboardHandler := handlers.NewDashboardHandler(userRepo, testRepo, attemptRepo)
	testHandler := handlers.NewTestHandler(testRepo, attemptRepo, userRepo)
	adminHandler := handlers.NewAdminHandler(testRepo, userRepo)
	teacherHandler := handlers.NewTeacherHandler(testRepo, userRepo, attemptRepo)

	// Initialize auth middleware
	authMiddleware := auth.NewMiddleware(s.sessionStore)

	// Public routes
	s.router.Get("/", s.handleHome)
	s.router.Get("/login", authHandler.ShowLogin)
	s.router.Post("/login", authHandler.Login)
	s.router.Get("/register", authHandler.ShowRegister)
	s.router.Post("/register", authHandler.Register)

	// Protected routes - require authentication
	s.router.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		// Dashboard
		r.Get("/dashboard", dashboardHandler.ShowDashboard)
		r.Get("/logout", authHandler.Logout)
		r.Get("/history", testHandler.History)

		// Tests - student routes
		r.Get("/tests", testHandler.ListTests)
		r.Get("/test/start", testHandler.StartTest)
		r.Get("/test/take", testHandler.TakeTest)
		r.Post("/test/answer", testHandler.SubmitAnswer)
		r.Post("/test/submit", testHandler.SubmitTest)
		r.Get("/test/results", testHandler.ViewResults)

		// Admin/Teacher routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireRole("admin", "teacher"))
			r.Get("/admin", adminHandler.ShowAdmin)
			r.Get("/admin/wizard", adminHandler.ShowWizard)
			r.Post("/admin/wizard", adminHandler.CreateWizardTest)
			r.Post("/admin/upload", adminHandler.UploadTest)

			// Teacher-specific routes
			r.Get("/teacher/dashboard", teacherHandler.ShowDashboard)
			r.Get("/teacher/upload", teacherHandler.ShowUpload)
			r.Post("/teacher/upload", teacherHandler.UploadTest)
			r.Get("/teacher/test/create", teacherHandler.ShowCreateTest)
			r.Post("/teacher/test/create", teacherHandler.CreateTest)
			r.Get("/teacher/test/{id}/edit", teacherHandler.EditTest)
			r.Post("/teacher/test/{id}/update", teacherHandler.UpdateTest)
			r.Get("/teacher/test/{id}/preview", teacherHandler.PreviewTest)
			r.Post("/teacher/test/{id}/publish", teacherHandler.PublishTest)
			r.Post("/teacher/test/{id}/unpublish", teacherHandler.UnpublishTest)
			r.Post("/teacher/test/{id}/delete", teacherHandler.DeleteTest)
			r.Delete("/teacher/test/{id}", teacherHandler.DeleteTest)

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireRole("admin"))
				r.Get("/admin/manage", adminHandler.ShowManagement)
				r.Post("/admin/manage/subjects", adminHandler.CreateSubject)
				r.Delete("/admin/manage/subjects/{id}", adminHandler.DeleteSubject)
				r.Post("/admin/test/{id}/delete", adminHandler.DeleteTest)
				r.Delete("/admin/test/{id}", adminHandler.DeleteTest)
				r.Post("/admin/test/{id}/update", adminHandler.UpdateTest)
				r.Post("/admin/test/{id}/remove-notes", adminHandler.RemoveTestNotes)

				// User management routes
				r.Get("/admin/users", adminHandler.ShowUserManagement)
				r.Post("/admin/users/create", adminHandler.CreateUser)
				r.Post("/admin/users/{id}/role", adminHandler.UpdateUserRole)
				r.Post("/admin/users/{id}/reset-password", adminHandler.ResetUserPassword)
				r.Post("/admin/users/{id}/delete", adminHandler.DeleteUser)
				r.Delete("/admin/users/{id}", adminHandler.DeleteUser)
			})
		})

		// Notes viewing route (for students and teachers)
		r.Get("/tests/{id}/notes", adminHandler.ServeTestNotes)
	})

	return s
}

// Router returns the configured Chi router.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// handleHome renders the home page.
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/layout.html", "views/home.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Session": nil,
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
