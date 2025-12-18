// Package handlers implements the HTTP handlers.
package handlers

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/SanthoshCheemala/FLARE/backend/internal/auth"
	"github.com/SanthoshCheemala/FLARE/backend/internal/client"
	"github.com/SanthoshCheemala/FLARE/backend/internal/config"
	"github.com/SanthoshCheemala/FLARE/backend/internal/jobs"
	"github.com/SanthoshCheemala/FLARE/backend/internal/models"
	"github.com/SanthoshCheemala/FLARE/backend/internal/psiadapter"
	"github.com/SanthoshCheemala/FLARE/backend/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Handler struct {
	repo       *repository.Repository
	jobManager *jobs.Manager
	psi        *psiadapter.Adapter
	psiClient  *client.PSIClient
	auth       *auth.Service
}

func NewHandler(repo *repository.Repository, jobManager *jobs.Manager, cfg *config.Config, authSvc *auth.Service) *Handler {
	// Initialize PSI client pointing to the remote server
	// In a real app, this URL would come from config
	psiClient := client.NewPSIClient("http://localhost:8081")

	return &Handler{
		repo:       repo,
		jobManager: jobManager,
		psi:        psiadapter.NewAdapter(cfg.PSI.MaxWorkers),
		psiClient:  psiClient,
		auth:       authSvc,
	}
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Login error for %s: %v", req.Email, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil || !auth.CheckPassword(req.Password, user.PasswordHash) {
		// Use constant time comparison to prevent timing attacks (CheckPassword does this)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.Active {
		http.Error(w, "Account inactive", http.StatusForbidden)
		return
	}

	// Update last login
	h.repo.UpdateUserLastLogin(r.Context(), user.ID)

	accessToken, err := h.auth.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.auth.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
		User:         *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UploadCustomerList handles uploading a new customer list CSV
func (h *Handler) UploadCustomerList(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
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
	description := r.FormValue("description")
	if name == "" {
		name = fmt.Sprintf("Upload %s", time.Now().Format("2006-01-02 15:04"))
	}

	// Save file to disk instead of DB
	uploadDir := "./data/uploads"
	if err := os.MkdirAll(uploadDir, 0700); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Generate temp file name
	fileName := fmt.Sprintf("customers_%d.csv", time.Now().UnixNano())
	finalPath := filepath.Join(uploadDir, fileName)

	// Log the absolute path for debugging
	if absPath, err := filepath.Abs(finalPath); err == nil {
		log.Printf("Saving customer upload to: %s", absPath)
	}

	dst, err := os.Create(finalPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Reset file pointer to beginning
	file.Seek(0, 0)
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	// Convert to absolute path for storage
	absPath, err := filepath.Abs(finalPath)
	if err != nil {
		absPath = finalPath
	}

	// Create list with file path (no user tracking)
	listID, err := h.repo.CreateCustomerList(r.Context(), name, description, absPath, 0)
	if err != nil {
		log.Printf("Error creating customer list in DB: %v", err)
		os.Remove(finalPath) // Cleanup
		http.Error(w, fmt.Sprintf("Failed to create list: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Created customer list ID %d with file path: %s", listID, absPath)

	// We no longer parse and insert records into the DB here to save time.
	// The records will be read directly from the CSV during screening.
	
	// Count lines for response (using CSV reader for accuracy)
	count := 0
	if csvFile, err := os.Open(finalPath); err == nil {
		defer csvFile.Close()
		reader := csv.NewReader(csvFile)
		// Skip header
		if _, err := reader.Read(); err == nil {
			// Count remaining records
			for {
				if _, err := reader.Read(); err == io.EOF {
					break
				} else if err == nil {
					count++
				}
			}
		}
	}
	
	// Update the record count in the database
	err = h.repo.UpdateCustomerListRecordCount(r.Context(), listID, count)
	if err != nil {
		log.Printf("Warning: failed to update record count: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    listID,
		"count": count,
	})
}

// UploadSanctionList handles uploading a new sanction list CSV
func (h *Handler) UploadSanctionList(w http.ResponseWriter, r *http.Request) {
	// In distributed mode, Client cannot upload sanctions.
	http.Error(w, "Sanction upload is only allowed on the Sanctions Authority Server", http.StatusForbidden)
}

// GetCustomerLists returns available customer lists
func (h *Handler) GetCustomerLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.repo.GetCustomerLists(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

// GetCustomerListHeaders returns headers for a customer list CSV
func (h *Handler) GetCustomerListHeaders(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	lists, err := h.repo.GetCustomerLists(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var filePath string
	for _, l := range lists {
		if l.ID == id {
			filePath = l.FilePath
			break
		}
	}

	if filePath == "" {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		http.Error(w, "Failed to read CSV headers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"headers": headers,
	})
}

// DeleteCustomerList deletes a customer list
func (h *Handler) DeleteCustomerList(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteCustomerList(r.Context(), id); err != nil {
		log.Printf("Failed to delete customer list: %v", err)
		http.Error(w, "Failed to delete customer list", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GetSanctionLists returns available sanction lists from the Server
func (h *Handler) GetSanctionLists(w http.ResponseWriter, r *http.Request) {
	// Fetch from remote server
	lists, err := h.psiClient.GetSanctionLists(r.Context())
	if err != nil {
		log.Printf("Failed to fetch sanction lists from server: %v", err)
		http.Error(w, "Failed to fetch sanction lists", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

// DeleteSanctionList deletes a sanction list (proxies to server)
func (h *Handler) DeleteSanctionList(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	if err := h.psiClient.DeleteSanctionList(r.Context(), id); err != nil {
		log.Printf("Failed to delete sanction list: %v", err)
		http.Error(w, "Failed to delete sanction list", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// StartScreening initiates a new screening job
func (h *Handler) StartScreening(w http.ResponseWriter, r *http.Request) {
	var req models.StartScreeningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate job ID
	jobID := fmt.Sprintf("screening_%d", time.Now().UnixNano())

	// Create screening job (no user tracking)
	job := h.jobManager.Create(jobID, req.Name, req.CustomerListID, req.SanctionListIDs, 0)

	// Create screening record
	screening := &models.Screening{
		JobID:           job.ID,
		Name:            req.Name,
		CustomerListID:  req.CustomerListID,
		SanctionListIDs: req.SanctionListIDs,
		Status:          "PENDING",
		CreatedBy:       0,
	}

	if err := h.repo.CreateScreening(r.Context(), screening); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create screening: %v", err), http.StatusInternalServerError)
		return
	}

	// Start screening in background - pass screening ID and mapping
	go h.runScreening(job, screening.ID, req.ColumnMapping)

	resp := models.StartScreeningResponse{
		JobID: job.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// runScreening executes the PSI screening process
func (h *Handler) runScreening(job *jobs.ScreeningJob, screeningID int64, columnMapping map[string]string) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Screening panic: %v\nStack: %s", r, debug.Stack())
			job.SetStatus(jobs.StatusFailed)
		}
	}()

	log.Printf("Starting screening job %s (ID: %d)", job.ID, screeningID)
	job.SetStatus(jobs.StatusRunning)

	// Initialize performance monitor
	perfMonitor := h.psi.NewPerformanceMonitor()
	
	// Helper function to get CPU usage (simplified)
	getCPUUsage := func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		// Rough CPU estimate based on GC activity and goroutines
		return float64(runtime.NumGoroutine()) * 2.5
	}

	// Stage 1: Preparing data
	job.AddProgress(jobs.PhaseServerInit, 10, "Loading customer and sanction data", nil)
	time.Sleep(500 * time.Millisecond)

	// Determine enabled columns from mapping
	var enabledColumns []string
	if columnMapping != nil {
		// We use a fixed order for consistency: name, dob, country, program
		// Check which ones are mapped
		if _, ok := columnMapping["name"]; ok && columnMapping["name"] != "" {
			enabledColumns = append(enabledColumns, "name")
		}
		if _, ok := columnMapping["dob"]; ok && columnMapping["dob"] != "" {
			enabledColumns = append(enabledColumns, "dob")
		}
		if _, ok := columnMapping["country"]; ok && columnMapping["country"] != "" {
			enabledColumns = append(enabledColumns, "country")
		}
	}
	// If empty, default to standard set
	if len(enabledColumns) == 0 {
		enabledColumns = []string{"name", "dob", "country"}
	}

	// Load data from CSV directly
	customerRecords, customerData, err := h.loadCustomerDataFromCSV(job.CustomerListID, columnMapping, enabledColumns)
	if err != nil {
		job.SetError(err)
		job.SetStatus(jobs.StatusFailed)
		return
	}

	// In distributed mode, we don't have sanction data locally
	job.SetCounts(len(customerData), 0)
	job.AddProgress(jobs.PhaseServerInit, 20, fmt.Sprintf("Loaded %d customers", len(customerData)), nil)

	// Log first few entries for debugging
	if len(customerData) > 0 {
		log.Printf("Sample customer data (first 3): %v", customerData[:min(3, len(customerData))])
		custHashes := psiadapter.HashDataPoints(customerData)
		log.Printf("Sample customer hashes: %v", custHashes[:min(3, len(custHashes))])
	}

	// Stage 2: Initializing session with remote server
	job.AddProgress(jobs.PhaseServerInit, 10, "Connecting to Sanctions Authority...", nil)
	time.Sleep(500 * time.Millisecond)

	// Convert list IDs to strings
	sanctionListIDs := make([]string, len(job.SanctionListIDs))
	for i, id := range job.SanctionListIDs {
		sanctionListIDs[i] = fmt.Sprintf("%d", id)
	}

	// Call Server to init session
	sessionID, serializedParams, err := h.psiClient.InitSession(ctx, sanctionListIDs, enabledColumns)
	if err != nil {
		job.SetError(fmt.Errorf("failed to init session with server: %w", err))
		job.SetStatus(jobs.StatusFailed)
		return
	}

	job.AddProgress(jobs.PhaseServerInit, 40, "Received public parameters from server", nil)

	// Deserialize params
	pp, msg, le, err := h.psi.DeserializeParams(serializedParams)
	if err != nil {
		job.SetError(fmt.Errorf("failed to deserialize params: %w", err))
		job.SetStatus(jobs.StatusFailed)
		return
	}

	// Construct a temporary ServerContext for encryption (we only need PP, Msg, LE)
	serverCtx := &psiadapter.ServerContext{
		PP:  pp,
		Msg: msg,
		LE:  le,
	}

	// Stage 3: Encrypting client data
	job.AddProgress(jobs.PhaseClientEncrypt, 30, "Generating client keys and encrypting dataset...", nil)
	time.Sleep(800 * time.Millisecond)

	ciphertexts, err := h.psi.EncryptClient(ctx, customerData, serverCtx)
	if err != nil {
		job.SetError(fmt.Errorf("failed to encrypt client data: %w", err))
		job.SetStatus(jobs.StatusFailed)
		return
	}

	// Get performance metrics after encryption
	metrics := perfMonitor.GetMetrics()
	memStats := perfMonitor.GetMemoryUsage()
	
	throughput := float64(0)
	memory := float64(0)
	cpu := getCPUUsage()
	
	if thr, ok := metrics["throughput_ops_per_sec"].(float64); ok {
		throughput = thr
	}
	if mem, ok := memStats["alloc_mb"].(float64); ok {
		memory = mem
	}

	job.AddProgress(jobs.PhaseClientEncrypt, 60, fmt.Sprintf("Encrypted %d records", len(ciphertexts)), map[string]string{
		"encrypted_records": fmt.Sprintf("%d", len(ciphertexts)),
		"throughput":        fmt.Sprintf("%.2f", throughput),
		"memory":            fmt.Sprintf("%.2f", memory),
		"cpu":               fmt.Sprintf("%.1f", cpu),
	})

	// Stage 4: Computing intersection (Remote)
	job.AddProgress(jobs.PhaseIntersection, 70, "Sending encrypted data to server for intersection...", nil)
	time.Sleep(1 * time.Second)

	// Log number of ciphertexts
	log.Printf("Number of ciphertexts: %d", len(ciphertexts))

	// Run intersection in background with heartbeat
	type intersectResult struct {
		matches []uint64
		err     error
	}
	resultChan := make(chan intersectResult, 1)

	go func() {
		matches, err := h.psiClient.Intersect(ctx, sessionID, ciphertexts)
		resultChan <- intersectResult{matches: matches, err: err}
	}()

	// Wait for result with heartbeat
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var matches []uint64

Loop:
	for {
		select {
		case res := <-resultChan:
			if res.err != nil {
				job.SetError(res.err)
				job.SetStatus(jobs.StatusFailed)
				return
			}
			matches = res.matches
			break Loop
		case <-ticker.C:
			// Send heartbeat with updated metrics
			metrics := perfMonitor.GetMetrics()
			memStats := perfMonitor.GetMemoryUsage()
			
			throughput := float64(0)
			memory := float64(0)
			cpu := getCPUUsage()
			
			if thr, ok := metrics["throughput_ops_per_sec"].(float64); ok {
				throughput = thr
			}
			if mem, ok := memStats["alloc_mb"].(float64); ok {
				memory = mem
			}
			
			job.AddProgress(jobs.PhaseIntersection, 75, "Intersecting... (this may take a few minutes)", map[string]string{
				"throughput": fmt.Sprintf("%.2f", throughput),
				"memory":     fmt.Sprintf("%.2f", memory),
				"cpu":        fmt.Sprintf("%.1f", cpu),
			})
		}
	}

	log.Printf("PSI returned matches: %v", matches)
	
	// Fallback removed for security. If PSI returns 0 matches but we expect some,
	// it means either no intersection exists or the PSI protocol failed.
	// We trust the crypto result.
	if len(matches) == 0 {
		log.Printf("PSI returned 0 matches. This could be correct, or due to data mismatch.")
	}
	
	// Get final performance metrics
	finalMetrics := perfMonitor.GetMetrics()
	finalMemStats := perfMonitor.GetMemoryUsage()
	
	finalThroughput := float64(0)
	finalMemory := float64(0)
	finalCPU := getCPUUsage()
	
	if thr, ok := finalMetrics["throughput_ops_per_sec"].(float64); ok {
		finalThroughput = thr
	}
	if mem, ok := finalMemStats["alloc_mb"].(float64); ok {
		finalMemory = mem
	}
	
	job.AddProgress(jobs.PhaseIntersection, 85, fmt.Sprintf("Found %d potential matches", len(matches)), map[string]string{
		"potential_matches": fmt.Sprintf("%d", len(matches)),
		"throughput":        fmt.Sprintf("%.2f", finalThroughput),
		"memory":            fmt.Sprintf("%.2f", finalMemory),
		"cpu":               fmt.Sprintf("%.1f", finalCPU),
	})

	// Stage 5: Storing results
	job.AddProgress(jobs.PhasePersist, 90, "Saving results to database", nil)

	// Resolve matches using in-memory maps
	var resultIDs []int64
	
	// Create a map of hash -> customer record
	customerMap := make(map[int64]*models.Customer)
	for i, hash := range psiadapter.HashDataPoints(customerData) {
		customerMap[int64(hash)] = customerRecords[i]
	}

	log.Printf("PSI returned %d match hashes", len(matches))
	log.Printf("Customer map has %d entries", len(customerMap))

	// Fetch matched sanctions from SERVER (distributed mode)
	sanctionRecords, err := h.psiClient.ResolveSanctions(ctx, sessionID, matches)
	if err != nil {
		log.Printf("Failed to resolve sanctions from server: %v", err)
		job.SetError(fmt.Errorf("failed to resolve sanctions: %w", err))
		job.SetStatus(jobs.StatusFailed)
		return
	}

	// Create sanction hash map
	sanctionMap := make(map[int64]*models.Sanction)
	for i := range sanctionRecords {
		sanctionMap[sanctionRecords[i].Hash] = sanctionRecords[i]
	}
	log.Printf("Resolved %d sanctions from server", len(sanctionMap))

	for _, matchHash := range matches {
		customer, cOk := customerMap[int64(matchHash)]
		sanction, sOk := sanctionMap[int64(matchHash)]
		
		log.Printf("Processing match hash %d: customer found=%v, sanction found=%v", matchHash, cOk, sOk)
		
		if cOk && sOk {
			log.Printf("Match found: Customer=%s (%s, %s) <-> Sanction=%s (%s, %s, %s)", 
				customer.Name, customer.DOB, customer.Country,
				sanction.Name, sanction.DOB, sanction.Country, sanction.Program)
			
			// Ensure customer is in database (for client-side CSVs, they aren't inserted initially)
			if customer.ID == 0 {
				customer.Hash = int64(matchHash) // Ensure hash is set
				if err := h.repo.CreateCustomer(ctx, customer); err != nil {
					log.Printf("Warning: Failed to save customer to local DB: %v", err)
					// We can't save the result without a customer ID
					continue
				}
				log.Printf("Inserted matched customer %s with ID %d", customer.Name, customer.ID)
			}

			// Save sanction to database temporarily for result linking
			if err := h.repo.CreateSanction(ctx, sanction); err != nil {
				log.Printf("Warning: Failed to save sanction to local DB: %v", err)
				// Continue anyway - we just won't have a local copy
			}
			
			result := &models.ScreeningResult{
				ScreeningID: screeningID,
				CustomerID:  customer.ID,
				SanctionID:  sanction.ID,
				MatchScore:  1.0,
				Status:      "PENDING",
			}
			
			if err := h.repo.CreateScreeningResult(ctx, result); err != nil {
				log.Printf("Failed to save result: %v", err)
			} else {
				resultIDs = append(resultIDs, result.ID)
				log.Printf("Successfully saved screening result ID %d", result.ID)
			}
		} else {
			log.Printf("Warning: Match hash %d found but missing customer=%v or sanction=%v", matchHash, !cOk, !sOk)
		}
	}

	log.Printf("Total matches saved: %d", len(resultIDs))
	job.SetResults(resultIDs, len(resultIDs))

	// Update screening status
	h.repo.UpdateScreeningStatus(ctx, job.ID, "COMPLETED", len(resultIDs))

	job.AddProgress(jobs.PhaseComplete, 100, fmt.Sprintf("Screening complete with %d matches", len(resultIDs)), map[string]string{
		"final_matches": fmt.Sprintf("%d", len(resultIDs)),
	})
	job.SetStatus(jobs.StatusCompleted)
}

// Helper functions to load data from CSV
func (h *Handler) loadCustomerDataFromCSV(listID int64, mapping map[string]string, enabledColumns []string) ([]*models.Customer, []string, error) {
	// Get list metadata to find file path
	lists, err := h.repo.GetCustomerLists(context.Background())
	if err != nil {
		return nil, nil, err
	}
	
	var filePath string
	for _, l := range lists {
		if l.ID == listID {
			filePath = l.FilePath
			break
		}
	}
	
	if filePath == "" {
		log.Printf("Warning: No file path found for customer list ID %d", listID)
		return nil, nil, fmt.Errorf("no file path found for customer list ID %d", listID)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}

	// Map headers
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}
	
	getValue := func(record []string, colName string) string {
		// Use mapping if provided
		if mapping != nil {
			if mappedCol, ok := mapping[colName]; ok {
				// mappedCol is the CSV header name from frontend
				if idx, ok := headerMap[strings.ToLower(strings.TrimSpace(mappedCol))]; ok && idx < len(record) {
					return record[idx]
				}
			}
		}

		// Fallback to auto-detection
		if idx, ok := headerMap[colName]; ok && idx < len(record) {
			return record[idx]
		}
		// Fallback for common variations
		if colName == "id" {
			if idx, ok := headerMap["customer_id"]; ok && idx < len(record) { return record[idx] }
		}
		if colName == "name" {
			if idx, ok := headerMap["full_name"]; ok && idx < len(record) { return record[idx] }
		}
		return ""
	}

	var records []*models.Customer
	var strings []string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		customer := &models.Customer{
			ExternalID: getValue(record, "id"),
			Name:       getValue(record, "name"),
			DOB:        getValue(record, "dob"),
			Country:    getValue(record, "country"),
			ListID:     listID,
		}
		
		if customer.Name == "" && len(record) >= 2 {
			customer.Name = record[1]
		}

		records = append(records, customer)
		
		// Use dynamic serialization
		vals := map[string]string{
			"name":    customer.Name,
			"dob":     customer.DOB,
			"country": customer.Country,
		}
		strings = append(strings, psiadapter.SerializeDynamic(vals, enabledColumns))
	}

	return records, strings, nil
}

func (h *Handler) loadSanctionDataFromCSV(listIDs []int64) ([]*models.Sanction, []string, error) {
	var allRecords []*models.Sanction
	var allStrings []string

	// Get all lists to find paths
	lists, err := h.repo.GetSanctionLists(context.Background())
	if err != nil {
		return nil, nil, err
	}
	
	listMap := make(map[int64]string)
	for _, l := range lists {
		listMap[l.ID] = l.FilePath
	}

	for _, listID := range listIDs {
		filePath := listMap[listID]
		if filePath == "" {
			log.Printf("Warning: No file path found for sanction list ID %d, skipping", listID)
			continue
		}

		file, err := os.Open(filePath)
		if err != nil {
			// Skip missing files or handle error
			log.Printf("Warning: could not open sanction file %s: %v", filePath, err)
			continue
		}
		defer file.Close()

		reader := csv.NewReader(file)
		headers, err := reader.Read()
		if err != nil {
			log.Printf("Warning: could not read headers from %s: %v", filePath, err)
			continue
		}

		// Map headers
		headerMap := make(map[string]int)
		for i, h := range headers {
			headerMap[h] = i
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

			sanction := &models.Sanction{
				Name:    getValue(record, "name"),
				DOB:     getValue(record, "dob"),
				Country: getValue(record, "country"),
				ListID:  listID,
			}
			
			program := getValue(record, "sanction_program")
			if program == "" {
				program = getValue(record, "program")
			}

			allRecords = append(allRecords, sanction)
			allStrings = append(allStrings, psiadapter.SerializeSanction(sanction.Name, sanction.DOB, sanction.Country, program))
		}
	}

	return allRecords, allStrings, nil
}

// ScreeningStatus returns the current status of a screening job
func (h *Handler) ScreeningStatus(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		http.Error(w, "Missing jobId parameter", http.StatusBadRequest)
		return
	}

	job := h.jobManager.Get(jobID)
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	snapshot := job.GetSnapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&snapshot)
}

// ScreeningEvents streams real-time progress via Server-Sent Events
func (h *Handler) ScreeningEvents(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		http.Error(w, "Missing jobId parameter", http.StatusBadRequest)
		return
	}

	job := h.jobManager.Get(jobID)
	if job == nil {
		log.Printf("SSE connection failed: Job %s not found", jobID)
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Set SSE headers FIRST before checking flusher
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("SSE connection failed: ResponseWriter does not support flushing")
		// Cannot call http.Error here as headers are already written
		fmt.Fprintf(w, "event: error\ndata: Streaming unsupported\n\n")
		return
	}

	// Send initial connection message to keep connection alive
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Subscribe to job progress
	progressChan := job.Subscribe()
	defer job.Unsubscribe(progressChan)

	// Check if job is already done
	snapshot := job.GetSnapshot()
	if snapshot.Status == jobs.StatusCompleted || snapshot.Status == jobs.StatusFailed {
		// Send all past progress events
		for _, p := range snapshot.Progress {
			data, _ := json.Marshal(p)
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		flusher.Flush()
		// Send a final event to signal completion
		fmt.Fprintf(w, "event: done\ndata: Job completed\n\n")
		flusher.Flush()
		return
	}

	// Stream progress events
	for {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// Channel closed, job is done - send completion event
				fmt.Fprintf(w, "event: done\ndata: Job completed\n\n")
				flusher.Flush()
				return
			}

			data, _ := json.Marshal(progress)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			
			// If this is the final progress event, send completion signal
			if progress.Phase == jobs.PhaseComplete || progress.Phase == "failed" {
				fmt.Fprintf(w, "event: done\ndata: Job completed\n\n")
				flusher.Flush()
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}

// GetScreeningResults returns paginated screening results
func (h *Handler) GetScreeningResults(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		http.Error(w, "Missing jobId parameter", http.StatusBadRequest)
		return
	}

	// Parse pagination params
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	// Query results directly from database
	results, err := h.repo.GetScreeningResultsByJobID(r.Context(), jobID, limit, offset)
	if err != nil {
		log.Printf("Error fetching screening results for job %s: %v", jobID, err)
		http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
		return
	}

	// Get total count
	totalCount, err := h.repo.CountScreeningResultsByJobID(r.Context(), jobID)
	if err != nil {
		totalCount = int64(len(results))
	}

	response := map[string]interface{}{
		"results": results,
		"total":   totalCount,
		"limit":   limit,
		"offset":  offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateResultStatus updates the status of a screening result
func (h *Handler) UpdateResultStatus(w http.ResponseWriter, r *http.Request) {
	resultIDStr := chi.URLParam(r, "resultId")
	if resultIDStr == "" {
		http.Error(w, "Missing resultId parameter", http.StatusBadRequest)
		return
	}

	resultID, err := strconv.ParseInt(resultIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid resultId", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"PENDING":        true,
		"CONFIRMED":      true,
		"FALSE_POSITIVE": true,
	}
	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	// Update in database
	if err := h.repo.UpdateResultStatus(r.Context(), resultID, req.Status); err != nil {
		log.Printf("Failed to update result status: %v", err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated result %d status to %s", resultID, req.Status)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      resultID,
		"status":  req.Status,
	})
}

// GetStats returns dashboard statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	totalScreenings, totalMatches, activeLists, recentScreenings, err := h.repo.GetDashboardStats(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"totalScreenings":  totalScreenings,
		"totalMatches":     totalMatches,
		"activeLists":      activeLists,
		"recentScreenings": recentScreenings,
		"systemStatus":     "Healthy",
		"activeWorkers":    h.psi.GetWorkerCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetPerformanceMetrics returns real-time system performance metrics
func (h *Handler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory in MB
	allocMB := float64(m.Alloc) / 1024 / 1024
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	// Get the latest screening for metrics (if any)
	_, _, _, recentScreenings, _ := h.repo.GetDashboardStats(r.Context())
	
	// Default values - will be updated from last screening if available
	perfMetrics := map[string]interface{}{
		"total_time_seconds":        0.0,
		"total_time_formatted":      "0s",
		"key_gen_time_seconds":      0.0,
		"key_gen_time_formatted":    "0s",
		"key_gen_percent":           0.0,
		"hashing_time_seconds":      0.0,
		"hashing_time_formatted":    "0s",
		"hashing_percent":           0.0,
		"witness_time_seconds":      0.0,
		"witness_time_formatted":    "0s",
		"witness_percent":           0.0,
		"intersection_time_seconds": 0.0,
		"intersection_time_formatted": "0s",
		"intersection_percent":      0.0,
		"num_workers":               h.psi.GetWorkerCount(),
		"total_operations":          0,
		"throughput_ops_per_sec":    0.0,
	}

	// If we have recent screenings, estimate metrics based on last one
	if len(recentScreenings) > 0 && recentScreenings[0].Status == "COMPLETED" {
		lastScreening := recentScreenings[0]
		if !lastScreening.FinishedAt.IsZero() && !lastScreening.CreatedAt.IsZero() {
			duration := lastScreening.FinishedAt.Sub(lastScreening.CreatedAt).Seconds()
			if duration > 0 {
				perfMetrics["total_time_seconds"] = duration
				perfMetrics["total_time_formatted"] = fmt.Sprintf("%.2fs", duration)
				perfMetrics["total_operations"] = lastScreening.CustomerCount
				perfMetrics["throughput_ops_per_sec"] = float64(lastScreening.CustomerCount) / duration
				
				// Estimate phase breakdowns (typical PSI distribution)
				perfMetrics["key_gen_time_seconds"] = duration * 0.15
				perfMetrics["key_gen_time_formatted"] = fmt.Sprintf("%.2fs", duration * 0.15)
				perfMetrics["key_gen_percent"] = 15.0
				
				perfMetrics["hashing_time_seconds"] = duration * 0.10
				perfMetrics["hashing_time_formatted"] = fmt.Sprintf("%.2fs", duration * 0.10)
				perfMetrics["hashing_percent"] = 10.0
				
				perfMetrics["witness_time_seconds"] = duration * 0.25
				perfMetrics["witness_time_formatted"] = fmt.Sprintf("%.2fs", duration * 0.25)
				perfMetrics["witness_percent"] = 25.0
				
				perfMetrics["intersection_time_seconds"] = duration * 0.50
				perfMetrics["intersection_time_formatted"] = fmt.Sprintf("%.2fs", duration * 0.50)
				perfMetrics["intersection_percent"] = 50.0
			}
		}
	}

	memMetrics := map[string]interface{}{
		"alloc_mb":       allocMB,
		"total_alloc_mb": totalAllocMB,
		"sys_mb":         sysMB,
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
	}

	response := map[string]interface{}{
		"performance": perfMetrics,
		"memory":      memMetrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

// StreamLogs streams server logs via WebSocket
func (h *Handler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Open log file
	logFile, err := os.Open("server.log")
	if err != nil {
		// Try opening in current directory if path fails
		logFile, err = os.Open("./server.log")
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error opening log file: "+err.Error()))
			return
		}
	}
	defer logFile.Close()

	// Seek to end to tail new logs
	stat, err := logFile.Stat()
	if err == nil {
		startPos := stat.Size() - 2048
		if startPos < 0 {
			startPos = 0
		}
		logFile.Seek(startPos, io.SeekStart)
		
		// Read until end
		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			conn.WriteMessage(websocket.TextMessage, scanner.Bytes())
		}
	} else {
		logFile.Seek(0, io.SeekEnd)
	}

	reader := bufio.NewReader(logFile)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					conn.WriteMessage(websocket.TextMessage, []byte("Error reading log: "+err.Error()))
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
					return // Client disconnected
				}
			}
		case <-r.Context().Done():
			return
		}
	}
}
