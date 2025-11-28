// Package handlers implements the HTTP handlers.
package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/SanthoshCheemala/FLARE/backend/utils"
	"github.com/go-chi/chi/v5"
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

	// Parse CSV and insert records
	readFile, err := os.Open(finalPath)
	if err != nil {
		log.Printf("Failed to open saved file: %v", err)
	} else {
		defer readFile.Close()
		reader := csv.NewReader(readFile)
		headers, err := reader.Read()
		if err == nil {
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
				externalID := getValue(record, "id")
				if externalID == "" {
					externalID = fmt.Sprintf("EXT-%d", time.Now().UnixNano())
				}

				if name != "" {
					dp := models.PSIDataPoint{
						Name:    strings.ToLower(strings.TrimSpace(name)),
						DOB:     dob,
						Country: strings.ToLower(strings.TrimSpace(country)),
					}
					serialized, _ := utils.SerializeData(dp)
					hash := psiadapter.HashOne(serialized)

					customer := &models.Customer{
						ExternalID: externalID,
						Name:       name,
						DOB:        dob,
						Country:    country,
						Hash:       hash,
						ListID:     listID,
					}
					
					if err := h.repo.CreateCustomer(r.Context(), customer); err == nil {
						count++
					}
				}
			}
			
			// Update record count (we need to add UpdateCustomerListCount to repo if missing, or just rely on DB count)
			// For now, let's assume we can update it or just return the count
			log.Printf("Imported %d customers for list %d", count, listID)
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    listID,
				"count": count,
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    listID,
		"count": 0,
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

	// Start screening in background - pass screening ID
	go h.runScreening(job, screening.ID)

	resp := models.StartScreeningResponse{
		JobID: job.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// runScreening executes the PSI screening process
func (h *Handler) runScreening(job *jobs.ScreeningJob, screeningID int64) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Screening panic: %v\nStack: %s", r, debug.Stack())
			job.SetStatus(jobs.StatusFailed)
		}
	}()

	log.Printf("Starting screening job %s (ID: %d)", job.ID, screeningID)
	job.SetStatus(jobs.StatusRunning)

	// Stage 1: Preparing data
	job.AddProgress(jobs.PhaseServerInit, 10, "Loading customer and sanction data", nil)
	time.Sleep(500 * time.Millisecond)

	// Load data from DB (to get IDs)
	dbCustomers, err := h.repo.GetCustomersByListID(ctx, job.CustomerListID)
	if err != nil {
		job.SetError(err)
		job.SetStatus(jobs.StatusFailed)
		return
	}

	var customerRecords []*models.Customer
	var customerData []string
	for i := range dbCustomers {
		c := &dbCustomers[i]
		customerRecords = append(customerRecords, c)
		// Re-serialize for PSI (must match what was hashed during upload)
		dp := models.PSIDataPoint{
			Name:    strings.ToLower(strings.TrimSpace(c.Name)),
			DOB:     c.DOB,
			Country: strings.ToLower(strings.TrimSpace(c.Country)),
		}
		serialized, _ := utils.SerializeData(dp)
		customerData = append(customerData, serialized)
	}

	// In distributed mode, we don't have sanction data locally
	job.SetCounts(len(customerData), 0)
	job.AddProgress(jobs.PhaseServerInit, 20, fmt.Sprintf("Loaded %d customers", len(customerData)), nil)

	// Log first few entries for debugging
	if len(customerData) > 0 {
		log.Printf("Sample customer data (first 3): %v", customerData[:min(3, len(customerData))])
		custHashes := psiadapter.HashDataPoints(customerData[:min(3, len(customerData))])
		log.Printf("Sample customer hashes: %v", custHashes)
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
	sessionID, serializedParams, err := h.psiClient.InitSession(ctx, sanctionListIDs)
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

	job.AddProgress(jobs.PhaseClientEncrypt, 60, fmt.Sprintf("Encrypted %d records", len(ciphertexts)), map[string]string{
		"encrypted_records": fmt.Sprintf("%d", len(ciphertexts)),
	})

	// Stage 4: Computing intersection (Remote)
	job.AddProgress(jobs.PhaseIntersection, 70, "Sending encrypted data to server for intersection...", nil)
	time.Sleep(1 * time.Second)

	// Log number of ciphertexts
	log.Printf("Number of ciphertexts: %d", len(ciphertexts))

	matches, err := h.psiClient.Intersect(ctx, sessionID, ciphertexts)
	if err != nil {
		job.SetError(err)
		job.SetStatus(jobs.StatusFailed)
		return
	}

	log.Printf("PSI returned matches: %v", matches)
	
	log.Printf("PSI returned matches: %v", matches)
	
	// Fallback removed for security. If PSI returns 0 matches but we expect some,
	// it means either no intersection exists or the PSI protocol failed.
	// We trust the crypto result.
	if len(matches) == 0 {
		log.Printf("PSI returned 0 matches. This could be correct, or due to data mismatch.")
	}
	
	job.AddProgress(jobs.PhaseIntersection, 85, fmt.Sprintf("Found %d potential matches", len(matches)), map[string]string{
		"potential_matches": fmt.Sprintf("%d", len(matches)),
	})

	// Stage 5: Storing results
	job.AddProgress(jobs.PhasePersist, 90, "Saving results to database", nil)

	// Resolve matches using in-memory maps
	var resultIDs []int64
	
	// Create a map of hash -> customer record
	customerMap := make(map[uint64]*models.Customer)
	for i, hash := range psiadapter.HashDataPoints(customerData) {
		customerMap[hash] = customerRecords[i]
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
	sanctionMap := make(map[uint64]*models.Sanction)
	for i := range sanctionRecords {
		sanctionMap[sanctionRecords[i].Hash] = sanctionRecords[i]
	}
	log.Printf("Resolved %d sanctions from server", len(sanctionMap))

	for _, matchHash := range matches {
		customer, cOk := customerMap[matchHash]
		sanction, sOk := sanctionMap[matchHash]
		
		log.Printf("Processing match hash %d: customer found=%v, sanction found=%v", matchHash, cOk, sOk)
		
		if cOk && sOk {
			log.Printf("Match found: Customer=%s (%s, %s) <-> Sanction=%s (%s, %s, %s)", 
				customer.Name, customer.DOB, customer.Country,
				sanction.Name, sanction.DOB, sanction.Country, sanction.Program)
			
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
func (h *Handler) loadCustomerDataFromCSV(listID int64) ([]*models.Customer, []string, error) {
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
		strings = append(strings, psiadapter.SerializeCustomer(customer.Name, customer.DOB, customer.Country))
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
