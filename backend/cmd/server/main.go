package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/SanthoshCheemala/FLARE/backend/internal/config"
	"github.com/SanthoshCheemala/FLARE/backend/internal/models"
	"github.com/SanthoshCheemala/FLARE/backend/internal/psiadapter"
	"github.com/SanthoshCheemala/FLARE/backend/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

// SessionContext wraps ServerContext with additional metadata
type SessionContext struct {
	*psiadapter.ServerContext
	ListIDs []string // Sanction list IDs used in this session
}

type Server struct {
	router  *chi.Mux
	adapter *psiadapter.Adapter
	repo    *repository.Repository
	mu      sync.Mutex // Protects sessions map
	// Map of sessionID -> SessionContext
	sessions map[string]*SessionContext
}

func NewServer(repo *repository.Repository) *Server {
	s := &Server{
		router:   chi.NewRouter(),
		adapter:  psiadapter.NewAdapter(0), // Use all cores
		repo:     repo,
		sessions: make(map[string]*SessionContext),
	}
	
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(s.corsMiddleware)

	s.router.Get("/health", s.handleHealth)
	s.router.Get("/dashboard/stats", s.handleGetStats)

	s.router.Post("/session/init", s.handleInitSession)
	s.router.Post("/session/intersect", s.handleIntersect)
	s.router.Post("/session/{sessionID}/resolve", s.handleResolveSanctions)
	
	s.router.Get("/lists/sanctions", s.handleGetSanctions)
	s.router.Post("/lists/sanctions/upload", s.handleUploadSanctions)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("FLARE Server (Sanctions Authority) is running"))
}

type InitSessionRequest struct {
	SanctionListIDs []string `json:"sanctionListIds"` // IDs of lists to screen against
}

type InitSessionResponse struct {
	SessionID string                           `json:"sessionId"`
	Params    *psiadapter.SerializedServerParams `json:"params"`
}

func (s *Server) handleInitSession(w http.ResponseWriter, r *http.Request) {
	var req InitSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load real sanction data from DB
	sanctionData, err := s.loadSanctionData(req.SanctionListIDs)
	if err != nil {
		log.Printf("Failed to load sanction data: %v", err)
		http.Error(w, "Failed to load sanction data", http.StatusInternalServerError)
		return
	}

	// Initialize PSI Server Context
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	treePath := fmt.Sprintf("./data/server_trees/%s", sessionID)
	os.MkdirAll("./data/server_trees", 0755)

	serverCtx, err := s.adapter.InitServer(r.Context(), sanctionData, treePath)
	if err != nil {
		log.Printf("InitServer failed: %v", err)
		http.Error(w, "Failed to initialize server", http.StatusInternalServerError)
		return
	}

	// Serialize parameters
	serializedParams, err := s.adapter.SerializeParams(serverCtx)
	if err != nil {
		log.Printf("Failed to serialize params: %v", err)
		http.Error(w, "Failed to serialize parameters", http.StatusInternalServerError)
		return
	}

	// Store session with metadata (store the list IDs for later resolution)
	s.mu.Lock()
	s.sessions[sessionID] = &SessionContext{
		ServerContext: serverCtx,
		ListIDs:       req.SanctionListIDs,
	}
	s.mu.Unlock()

	resp := InitSessionResponse{
		SessionID: sessionID,
		Params:    serializedParams,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type IntersectRequest struct {
	SessionID   string                        `json:"sessionId"`
	Ciphertexts []psiadapter.ClientCiphertext `json:"ciphertexts"`
}

type IntersectResponse struct {
	Matches []uint64 `json:"matches"`
}

func (s *Server) handleIntersect(w http.ResponseWriter, r *http.Request) {
	// Note: ClientCiphertext might need custom unmarshaling if it's complex.
	// Assuming standard JSON works for now, or we might need to use Gob over HTTP body.
	// But let's try JSON first. If ClientCiphertext is a struct with exported fields, it should work.
	
	// Wait, ClientCiphertext is `psi.Cxtx`. We don't know if it has JSON tags.
	// If it fails, we'll need to use Gob for the request body.
	// Let's use Gob for the request body to be safe, as we did for params.
	
	// Actually, let's stick to JSON for the wrapper, but maybe base64 encode the gob-encoded ciphertexts?
	// Or just use Gob for the whole body.
	// Let's try standard JSON decode first.
	
	var req IntersectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionCtx, ok := s.sessions[req.SessionID]
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Unwrap ServerContext for PSI operation
	matches, err := s.adapter.DetectIntersection(r.Context(), sessionCtx.ServerContext, req.Ciphertexts)
	if err != nil {
		log.Printf("Intersection failed: %v", err)
		http.Error(w, "Intersection failed", http.StatusInternalServerError)
		return
	}

	resp := IntersectResponse{
		Matches: matches,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetSanctions(w http.ResponseWriter, r *http.Request) {
	lists, err := s.repo.GetSanctionLists(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func (s *Server) handleUploadSanctions(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	source := r.FormValue("source")
	description := r.FormValue("description")
	if name == "" {
		name = fmt.Sprintf("Sanctions %s", time.Now().Format("2006-01-02"))
	}

	uploadDir := "./data/server_uploads"
	if err := os.MkdirAll(uploadDir, 0700); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("sanctions_%d.csv", time.Now().UnixNano())
	finalPath := fmt.Sprintf("%s/%s", uploadDir, fileName)

	dst, err := os.Create(finalPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Write file
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close() // Close on error
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}
	dst.Close() // Explicitly close to flush buffers before reading back
	
	absPath, _ := filepath.Abs(finalPath)

	listID, err := s.repo.CreateSanctionList(r.Context(), name, source, description, absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create list: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse CSV and insert records
	readFile, err := os.Open(finalPath)
	if err != nil {
		log.Printf("Failed to open saved file: %v", err)
	} else {
		defer readFile.Close()
		reader := csv.NewReader(readFile)
		headers, err := reader.Read()
		if err == nil {
			log.Printf("CSV Headers found: %v", headers)
			headerMap := make(map[string]int)
			for i, h := range headers {
				headerMap[strings.ToLower(strings.TrimSpace(h))] = i
			}
			
			getValue := func(record []string, colName string) string {
				if idx, ok := headerMap[colName]; ok && idx < len(record) {
					return record[idx]
				}
				return ""
			}

			count := 0
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}

				name := getValue(record, "name")
				dob := getValue(record, "dob")
				country := getValue(record, "country")
				program := getValue(record, "sanction_program")
				if program == "" {
					program = getValue(record, "program")
				}

				if name != "" {
					sanction := &models.Sanction{
						Name:    name,
						DOB:     dob,
						Country: country,
						Program: program,
						Source:  source,
						ListID:  listID,
						Hash:    psiadapter.HashOne(psiadapter.SerializeSanction(name, dob, country, program)),
					}
					if err := s.repo.CreateSanction(r.Context(), sanction); err == nil {
						count++
					}
				}
			}
			
			// Update record count in database
			if err := s.repo.UpdateSanctionListCount(r.Context(), listID, count); err != nil {
				log.Printf("Failed to update list count: %v", err)
			}
			log.Printf("Imported %d sanctions for list %d", count, listID)
		} else {
			log.Printf("Failed to read CSV headers: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": listID,
	})
}

func (s *Server) loadSanctionData(listIDs []string) ([]string, error) {
	var ids []int64
	for _, idStr := range listIDs {
		var id int64
		fmt.Sscanf(idStr, "%d", &id)
		ids = append(ids, id)
	}
	
	var allStrings []string
	
	lists, err := s.repo.GetSanctionLists(context.Background())
	if err != nil {
		return nil, err
	}
	
	listMap := make(map[int64]string)
	for _, l := range lists {
		listMap[l.ID] = l.FilePath
	}

	for _, listID := range ids {
		filePath := listMap[listID]
		if filePath == "" {
			continue
		}

		file, err := os.Open(filePath)
		if err != nil {
			continue
		}
		defer file.Close()

		reader := csv.NewReader(file)
		headers, err := reader.Read()
		if err != nil {
			continue
		}

		headerMap := make(map[string]int)
		for i, h := range headers {
			headerMap[strings.ToLower(strings.TrimSpace(h))] = i
		}
		
		getValue := func(record []string, colName string) string {
			if idx, ok := headerMap[colName]; ok && idx < len(record) {
				return record[idx]
			}
			return ""
		}

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			name := getValue(record, "name")
			dob := getValue(record, "dob")
			country := getValue(record, "country")
			program := getValue(record, "sanction_program")
			if program == "" {
				program = getValue(record, "program")
			}

			allStrings = append(allStrings, psiadapter.SerializeSanction(name, dob, country, program))
		}
	}
	
	return allStrings, nil
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	// Server-specific stats
	lists, _ := s.repo.GetSanctionLists(r.Context())
	
	totalEntities := 0
	for _, list := range lists {
		totalEntities += list.RecordCount
	}
	
	stats := map[string]interface{}{
		"totalScreenings": 0, // Server doesn't track screenings
		"totalMatches":    0,
		"activeLists":     len(lists),
		"totalEntities":   totalEntities,
		"recentScreenings": []interface{}{},
		"systemStatus":    "OPERATIONAL",
		"activeWorkers":   8,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleResolveSanctions returns full sanction details for matched hashes
func (s *Server) handleResolveSanctions(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		http.Error(w, "Missing sessionID", http.StatusBadRequest)
		return
	}

	var req struct {
		Hashes []uint64 `json:"hashes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the session to find which sanction lists were used
	s.mu.Lock()
	serverCtx, exists := s.sessions[sessionID]
	s.mu.Unlock()

	if !exists {
		http.Error(w, "Session not found or expired", http.StatusNotFound)
		return
	}

	// Load all sanctions from the lists used in this session
	listIDs := make([]int64, len(serverCtx.ListIDs))
	for i, idStr := range serverCtx.ListIDs {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		listIDs[i] = id
	}

	sanctions, err := s.repo.GetSanctionsByListIDs(r.Context(), listIDs)
	if err != nil {
		log.Printf("Failed to load sanctions: %v", err)
		http.Error(w, "Failed to load sanctions", http.StatusInternalServerError)
		return
	}

	// Create hash map for O(1) lookup
	hashSet := make(map[uint64]bool)
	for _, hash := range req.Hashes {
		hashSet[hash] = true
	}

	// Filter sanctions that match the provided hashes
	var matchedSanctions []map[string]interface{}
	for _, sanction := range sanctions {
		if hashSet[sanction.Hash] {
			matchedSanctions = append(matchedSanctions, map[string]interface{}{
				"hash":    sanction.Hash,
				"name":    sanction.Name,
				"dob":     sanction.DOB,
				"country": sanction.Country,
				"program": sanction.Program,
				"source":  sanction.Source,
			})
		}
	}

	log.Printf("Resolved %d sanctions for session %s from %d hashes", len(matchedSanctions), sessionID, len(req.Hashes))

	resp := map[string]interface{}{
		"sanctions": matchedSanctions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	if port == "" {
		port = "8081"
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure data directory exists for SQLite
	if cfg.DatabaseDriver() == "sqlite3" {
		os.MkdirAll("./data", 0755)
	}

	// Use a separate DB for Server if needed, or same one for demo
	// For strict separation, we should use a different DB file, e.g., flare_server.db
	// But config loads from env. Let's override DSN for server if using sqlite.
	dsn := cfg.DatabaseDSN()
	if cfg.DatabaseDriver() == "sqlite3" {
		dsn = "./data/flare_server.db"
	}

	db, err := sql.Open(cfg.DatabaseDriver(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	repo := repository.New(db)
	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	server := NewServer(repo)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: server.router,
	}

	go func() {
		log.Printf("Starting FLARE Server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server stopped")
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
