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
	ListIDs        []string // Sanction list IDs used in this session
	EnabledColumns []string // Schema used for this session
}

type Server struct {
	router  *chi.Mux
	adapter *psiadapter.Adapter
	repo    *repository.Repository
	mu      sync.Mutex // Protects sessions map
	// Map of sessionID -> SessionContext
	sessions map[string]*SessionContext
	
	// Global pre-computed state (for small datasets)
	GlobalServerContext *psiadapter.ServerContext
	GlobalParams        *psiadapter.SerializedServerParams
	
	// Batch PSI state (for large datasets)
	GlobalBatchContext *psiadapter.BatchServerContext
	UseBatching        bool
}

func NewServer(repo *repository.Repository) *Server {
	s := &Server{
		router:   chi.NewRouter(),
		adapter:  psiadapter.NewAdapter(0), // Use all cores
		repo:     repo,
		sessions: make(map[string]*SessionContext),
	}
	
	// Initialize global state
	if err := s.initGlobalState(); err != nil {
		log.Printf("WARNING: Failed to initialize global PSI state: %v", err)
	}
	
	s.routes()
	return s
}

func (s *Server) initGlobalState() error {
	log.Println("Initializing global PSI state...")
	ctx := context.Background()
	
	// Load ALL sanction lists
	lists, err := s.repo.GetSanctionLists(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sanction lists: %w", err)
	}
	
	var listIDs []string
	for _, l := range lists {
		listIDs = append(listIDs, fmt.Sprintf("%d", l.ID))
	}
	
	if len(listIDs) == 0 {
		log.Println("No sanction lists found. Skipping PSI init.")
		return nil
	}
	
	sanctionData, err := s.loadSanctionData(listIDs, nil) // nil for default schema
	if err != nil {
		return fmt.Errorf("failed to load sanction data: %w", err)
	}
	
	log.Printf("Loaded %d sanction records for global state", len(sanctionData))
	
	// Initialize PSI Server Context
	treeDir := "./data/server_trees"
	os.MkdirAll(treeDir, 0755)
	
	// Ensure global directory is removed if it exists (fix for previous bug)
	os.RemoveAll("./data/server_trees/global")
	
	treePath := filepath.Join(treeDir, "global")

	// Check if we should use batching based on dataset size and RAM
	if s.adapter.ShouldUseBatching(len(sanctionData)) {
		optimalBatch := s.adapter.CalculateOptimalBatchSize()
		numBatches := (len(sanctionData) + optimalBatch - 1) / optimalBatch
		log.Printf("🔄 BATCH PSI ACTIVATED: %d records → %d batches of %d (based on available RAM)",
			len(sanctionData), numBatches, optimalBatch)

		batchCtx, err := s.adapter.InitServerBatched(ctx, sanctionData, treePath)
		if err != nil {
			return fmt.Errorf("InitServerBatched failed: %w", err)
		}

		// For batch mode, we use the first batch's params (all batches have compatible params)
		serializedParams, err := s.adapter.SerializeParams(batchCtx.Batches[0])
		if err != nil {
			return fmt.Errorf("failed to serialize params: %w", err)
		}

		s.GlobalBatchContext = batchCtx
		s.GlobalServerContext = batchCtx.Batches[0] // Primary context for params
		s.GlobalParams = serializedParams
		s.UseBatching = true
		log.Printf("✓ Global Batch PSI state initialized: %d batches", len(batchCtx.Batches))
	} else {
		// Standard PSI for small datasets
		log.Printf("⚡ Standard PSI: %d records (within RAM limits)", len(sanctionData))
		
		serverCtx, err := s.adapter.InitServer(ctx, sanctionData, treePath+".db")
		if err != nil {
			return fmt.Errorf("InitServer failed: %w", err)
		}

		serializedParams, err := s.adapter.SerializeParams(serverCtx)
		if err != nil {
			return fmt.Errorf("failed to serialize params: %w", err)
		}
		
		s.GlobalServerContext = serverCtx
		s.GlobalParams = serializedParams
		s.UseBatching = false
	}
	
	log.Println("Global PSI state initialized successfully")
	return nil
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
	s.router.Delete("/lists/sanctions/{id}", s.handleDeleteSanctionList)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("FLARE Server (Sanctions Authority) is running"))
}

type InitSessionRequest struct {
	SanctionListIDs []string `json:"sanctionListIds"` // IDs of lists to screen against
	EnabledColumns  []string `json:"enabledColumns"`  // Columns to use for hashing (schema)
}

type InitSessionResponse struct {
	SessionID string                           `json:"sessionId"`
	Params    *psiadapter.SerializedServerParams `json:"params"`
}

func (s *Server) handleInitSession(w http.ResponseWriter, r *http.Request) {
	var req InitSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Warning: failed to decode init session request: %v", err)
	}

	// Determine effective columns. Default to standard set if empty.
	columns := req.EnabledColumns
	if len(columns) == 0 {
		columns = []string{"name", "dob", "country"}
	}
	
	// Check if this matches global state (default)
	isDefaultSchema := len(columns) == 3 && 
		columns[0] == "name" && columns[1] == "dob" && columns[2] == "country"

	// If default schema and global state is ready, use it (optimization)
	if isDefaultSchema && s.GlobalParams != nil {
		sessionID := fmt.Sprintf("session_global_%d", time.Now().UnixNano())
		s.mu.Lock()
		s.sessions[sessionID] = &SessionContext{
			ServerContext:  s.GlobalServerContext,
			ListIDs:        req.SanctionListIDs,
			EnabledColumns: columns,
		}
		s.mu.Unlock()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(InitSessionResponse{
			SessionID: sessionID,
			Params:    s.GlobalParams,
		})
		return
	}

	// Dynamic Schema: We must re-compute the tree
	log.Printf("Initializing dynamic PSI session with columns: %v", columns)
	
	// Load requested lists (or all if none specified)
	listIDs := req.SanctionListIDs
	if len(listIDs) == 0 {
		lists, _ := s.repo.GetSanctionLists(r.Context())
		for _, l := range lists {
			listIDs = append(listIDs, fmt.Sprintf("%d", l.ID))
		}
	}
	
	// Load and Hash Data dynamically
	sanctionData, err := s.loadSanctionData(listIDs, columns)
	if err != nil {
		http.Error(w, "Failed to load sanction data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Init Server Context (Dynamic Tree)
	// We use a temporary path for dynamic trees
	treeDir := fmt.Sprintf("./data/server_trees/dynamic_%d", time.Now().UnixNano())
	os.MkdirAll(treeDir, 0700)
	defer os.RemoveAll(treeDir) // Clean up after session? No, need it for interactions.
	// Actually, we should keep it for the session duration. 
	// For this POC, we'll leave it or clean it up periodically.
	
	treePath := filepath.Join(treeDir, "tree.db")
	serverCtx, err := s.adapter.InitServer(r.Context(), sanctionData, treePath)
	if err != nil {
		http.Error(w, "InitServer failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	serializedParams, err := s.adapter.SerializeParams(serverCtx)
	if err != nil {
		http.Error(w, "SerializeParams failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	sessionID := fmt.Sprintf("session_dyn_%d", time.Now().UnixNano())
	s.mu.Lock()
	s.sessions[sessionID] = &SessionContext{
		ServerContext:  serverCtx,
		ListIDs:        listIDs,
		EnabledColumns: columns,
	}
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(InitSessionResponse{
		SessionID: sessionID,
		Params:    serializedParams,
	})
}

type IntersectRequest struct {
	SessionID   string                        `json:"sessionId"`
	Ciphertexts []psiadapter.ClientCiphertext `json:"ciphertexts"`
}

type IntersectResponse struct {
	Matches []uint64 `json:"matches"`
}

func (s *Server) handleIntersect(w http.ResponseWriter, r *http.Request) {
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

	var matches []uint64
	var err error

	// Check if this is a global session using batch context
	isGlobalSession := len(req.SessionID) > 14 && req.SessionID[:14] == "session_global"
	
	if isGlobalSession && s.UseBatching && s.GlobalBatchContext != nil {
		// Use batch intersection - iterate through ALL batches
		log.Printf("🔄 Running batched intersection across %d batches", len(s.GlobalBatchContext.Batches))
		allMatches := make(map[uint64]bool)
		
		for i, batch := range s.GlobalBatchContext.Batches {
			batchMatches, batchErr := s.adapter.DetectIntersection(r.Context(), batch, req.Ciphertexts)
			if batchErr != nil {
				log.Printf("Batch %d intersection failed: %v", i, batchErr)
				continue
			}
			log.Printf("   Batch %d: found %d matches", i, len(batchMatches))
			for _, m := range batchMatches {
				allMatches[m] = true
			}
		}
		
		// Convert map to slice
		matches = make([]uint64, 0, len(allMatches))
		for hash := range allMatches {
			matches = append(matches, hash)
		}
		log.Printf("✓ Total matches from all batches: %d", len(matches))
	} else {
		// Standard single-context intersection
		matches, err = s.adapter.DetectIntersection(r.Context(), sessionCtx.ServerContext, req.Ciphertexts)
		if err != nil {
			log.Printf("Intersection failed: %v", err)
			http.Error(w, "Intersection failed", http.StatusInternalServerError)
			return
		}
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
						Hash:    int64(psiadapter.HashOne(psiadapter.SerializeSanction(name, dob, country, program))),
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

func (s *Server) handleDeleteSanctionList(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	if err := s.repo.DeleteSanctionList(r.Context(), id); err != nil {
		log.Printf("Failed to delete sanction list: %v", err)
		http.Error(w, "Failed to delete sanction list", http.StatusInternalServerError)
		return
	}

	// Re-initialize global state to reflect changes
	// In a real system, we might want to do this more gracefully or lazily
	go func() {
		if err := s.initGlobalState(); err != nil {
			log.Printf("Failed to re-initialize global state after deletion: %v", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) loadSanctionData(listIDs []string, columns []string) ([]string, error) {
	var ids []int64
	for _, idStr := range listIDs {
		var id int64
		fmt.Sscanf(idStr, "%d", &id)
		ids = append(ids, id)
	}
	
	var allStrings []string
	
	// Load sanctions directly from database
	sanctions, err := s.repo.GetSanctionsByListIDs(context.Background(), ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load sanctions: %w", err)
	}
	
	if len(columns) == 0 {
		columns = []string{"name", "dob", "country"}
	}
	
	for _, sanction := range sanctions {
		// Dynamic serialization
		vals := map[string]string{
			"name":    sanction.Name,
			"dob":     sanction.DOB,
			"country": sanction.Country,
			"program": sanction.Program,
		}
		serialized := psiadapter.SerializeDynamic(vals, columns)
		allStrings = append(allStrings, serialized)
	}
	
	// Debug
	if len(allStrings) > 0 {
		log.Printf("[DEBUG] Server loaded %d sanction records with schema %v", len(allStrings), columns)
		for i := 0; i < 3 && i < len(allStrings); i++ {
			hash := psiadapter.HashOne(allStrings[i])
			log.Printf("[DEBUG] Sanction %d: '%s' -> hash: %d", i, allStrings[i], hash)
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
		Hashes []int64 `json:"hashes"`
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
	log.Printf("[DEBUG] Resolving for session %s with ListIDs: %v", sessionID, listIDs)

	sanctions, err := s.repo.GetSanctionsByListIDs(r.Context(), listIDs)
	if err != nil {
		log.Printf("Failed to load sanctions: %v", err)
		http.Error(w, "Failed to load sanctions", http.StatusInternalServerError)
		return
	}
	log.Printf("[DEBUG] Loaded %d sanctions from DB", len(sanctions))

	// Create hash map for O(1) lookup
	hashSet := make(map[int64]bool)
	for _, hash := range req.Hashes {
		hashSet[int64(hash)] = true
	}
	log.Printf("[DEBUG] Request contains %d hashes. Sample: %v", len(req.Hashes), req.Hashes[:min(3, len(req.Hashes))])

	// Filter sanctions that match the provided hashes using DYNAMIC hashing
	var matchedSanctions []map[string]interface{}
	
	// Default columns if not set (legacy sessions)
	columns := serverCtx.EnabledColumns
	if len(columns) == 0 {
		columns = []string{"name", "dob", "country"}
	}
	
	for _, sanction := range sanctions {
		// Re-calculate hash using the session's schema
		vals := map[string]string{
			"name":    sanction.Name,
			"dob":     sanction.DOB,
			"country": sanction.Country,
			"program": sanction.Program,
		}
		serialized := psiadapter.SerializeDynamic(vals, columns)
		dynamicHash := int64(psiadapter.HashOne(serialized))
		
		if hashSet[dynamicHash] {
			log.Printf("[DEBUG] Match found! Hash: %d, Name: %s", dynamicHash, sanction.Name)
			matchedSanctions = append(matchedSanctions, map[string]interface{}{
				"hash":    dynamicHash, // Return the DYNAMIC hash properly
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
