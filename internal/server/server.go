package server

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/database"
	"github.com/jchanning/gocase/internal/handlers"
	"github.com/jchanning/gocase/internal/llm"
	"github.com/jchanning/gocase/internal/repository"

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
func NewServer(db *database.Service, llmClient llm.QuestionGenerator) *Server {
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
	assignmentRepo := repository.NewAssignmentRepository(db.Pool())
	feedbackRepo := repository.NewFeedbackRepository(db.Pool())
	syllabusRepo := repository.NewSyllabusRepository(db.Pool())
	revisionRepo := repository.NewRevisionRepository(db.Pool())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, s.sessionStore)
	dashboardHandler := handlers.NewDashboardHandler(userRepo, attemptRepo, assignmentRepo)
	testHandler := handlers.NewTestHandler(testRepo, attemptRepo, userRepo, assignmentRepo, feedbackRepo)
	adminHandler := handlers.NewAdminHandler(testRepo, userRepo, llmClient)
	teacherHandler := handlers.NewTeacherHandler(testRepo, userRepo, attemptRepo, assignmentRepo, feedbackRepo, testRepo)
	manageHandler := handlers.NewManageHandler(testRepo, feedbackRepo)
	syllabusHandler := handlers.NewSyllabusHandler(syllabusRepo, testRepo)
	revisionHandler := handlers.NewRevisionHandler(syllabusRepo, revisionRepo)

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
		r.Get("/test/review", testHandler.ReviewTest)
		r.Post("/test/feedback/report", testHandler.ReportIssue)

		// Admin/Teacher routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireRole("admin", "teacher"))
			r.Get("/manage", manageHandler.ShowManage)
			r.Get("/admin", manageHandler.ShowManage)
			r.Get("/admin/manage", manageHandler.ShowManage)
			r.Post("/manage/test/{id}/submit-review", manageHandler.SubmitForReview)
			r.Post("/manage/test/{id}/approve", manageHandler.ApproveTest)
			r.Post("/manage/test/{id}/request-changes", manageHandler.RequestChanges)
			r.Post("/manage/feedback/{id}/update", manageHandler.UpdateFeedbackIssue)
			r.Get("/admin/wizard", adminHandler.ShowWizard)
			r.Post("/admin/wizard", adminHandler.CreateWizardTest)
			r.Post("/admin/upload", adminHandler.UploadTest)
			r.Get("/admin/generate", adminHandler.ShowGenerate)
			r.Post("/admin/generate", adminHandler.GenerateFromNotes)

			// Teacher-specific routes
			r.Get("/teacher/dashboard", teacherHandler.ShowDashboard)
			r.Get("/teacher/create", teacherHandler.ShowCreateTests)
			r.Get("/teacher/manage", teacherHandler.ShowManageTests)
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
			r.Get("/teacher/test/{id}/assign", teacherHandler.ShowAssignTest)
			r.Post("/teacher/test/{id}/assign", teacherHandler.AssignTest)

			// Syllabus management – admin and teacher
			r.Get("/admin/syllabus", syllabusHandler.List)
			r.Get("/admin/syllabus/new", syllabusHandler.ShowCreate)
			r.Post("/admin/syllabus", syllabusHandler.Create)
			r.Get("/admin/syllabus/{id}", syllabusHandler.ShowEdit)
			r.Post("/admin/syllabus/{id}", syllabusHandler.Update)
			r.Post("/admin/syllabus/{id}/publish", syllabusHandler.Publish)
			r.Post("/admin/syllabus/{id}/unpublish", syllabusHandler.Unpublish)
			r.Post("/admin/syllabus/{id}/section", syllabusHandler.AddSection)
			r.Post("/admin/syllabus/{id}/section/{sid}/update", syllabusHandler.UpdateSection)
			r.Post("/admin/syllabus/{id}/section/{sid}/delete", syllabusHandler.DeleteSection)
			r.Post("/admin/syllabus/{id}/topic", syllabusHandler.AddTopic)
			r.Post("/admin/syllabus/{id}/topic/{tid}/update", syllabusHandler.UpdateTopic)
			r.Post("/admin/syllabus/{id}/topic/{tid}/delete", syllabusHandler.DeleteTopic)
			r.Post("/admin/syllabus/{id}/topic/{tid}/tests", syllabusHandler.LinkTest)
			r.Post("/admin/syllabus/{id}/topic/{tid}/tests/{testid}/delete", syllabusHandler.UnlinkTest)

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireRole("admin"))
				r.Post("/admin/manage/subjects", adminHandler.CreateSubject)
				r.Delete("/admin/manage/subjects/{id}", adminHandler.DeleteSubject)
				r.Get("/admin/test/{id}/edit", adminHandler.EditTest)
				r.Get("/admin/test/{id}/preview", teacherHandler.PreviewTest)
				r.Get("/admin/test/{id}/pdf", adminHandler.ExportTestPDF)
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

		// Revision planner – student routes (all authenticated users)
		r.Get("/revision", revisionHandler.ShowRevision)
		r.Get("/revision/plan/new", revisionHandler.ShowCreatePlan)
		r.Post("/revision/plan", revisionHandler.CreatePlan)
		r.Get("/revision/plan/{id}", revisionHandler.ShowPlan)
		r.Post("/revision/session/{id}/complete", revisionHandler.CompleteSession)
		r.Post("/revision/session/{id}/skip", revisionHandler.SkipSession)
		r.Post("/revision/plan/{id}/delete", revisionHandler.DeletePlan)

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
