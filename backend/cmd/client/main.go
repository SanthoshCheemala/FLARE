package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SanthoshCheemala/FLARE/backend/internal/config"
	"github.com/SanthoshCheemala/FLARE/backend/internal/handlers"
	"github.com/SanthoshCheemala/FLARE/backend/internal/jobs"
	"github.com/SanthoshCheemala/FLARE/backend/internal/middleware"
	"github.com/SanthoshCheemala/FLARE/backend/internal/repository"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure data directory exists for SQLite
	if cfg.DatabaseDriver() == "sqlite3" {
		os.MkdirAll("./data", 0755)
	}

	db, err := sql.Open(cfg.DatabaseDriver(), cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	repo := repository.New(db)
	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}
	jobManager := jobs.NewManager(cfg.PSI.MaxScreenings)
	handler := handlers.NewHandler(repo, jobManager, cfg, nil)

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS([]string{"http://localhost:3000", "*"}))

	// WebSocket endpoint (must be outside Timeout middleware)
	r.Get("/ws/logs", handler.StreamLogs)

	// API endpoints with timeout
	r.Group(func(r chi.Router) {
		r.Use(chimiddleware.Timeout(60 * time.Second))

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// All endpoints are now public (no auth required)
		r.Post("/lists/customers/upload", handler.UploadCustomerList)
		r.Post("/lists/sanctions/upload", handler.UploadSanctionList)
		r.Get("/lists/customers", handler.GetCustomerLists)
		r.Get("/lists/customers/{id}/headers", handler.GetCustomerListHeaders)
		r.Delete("/lists/customers/{id}", handler.DeleteCustomerList)
		r.Get("/lists/sanctions", handler.GetSanctionLists)
		r.Delete("/lists/sanctions/{id}", handler.DeleteSanctionList)

		r.Post("/screenings", handler.StartScreening)
		r.Get("/screenings/{jobId}/status", handler.ScreeningStatus)
		r.Get("/screenings/{jobId}/events", handler.ScreeningEvents)
		r.Get("/screenings/{jobId}/results", handler.GetScreeningResults)
		
		r.Patch("/results/{resultId}/status", handler.UpdateResultStatus)
		
		r.Get("/dashboard/stats", handler.GetStats)
		r.Get("/performance/metrics", handler.GetPerformanceMetrics)
	})

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
