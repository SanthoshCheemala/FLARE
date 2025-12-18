package psiadapter

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/SanthoshCheemala/LE-PSI/pkg/LE"
	"github.com/SanthoshCheemala/LE-PSI/pkg/matrix"
	"github.com/SanthoshCheemala/LE-PSI/pkg/psi"
	"github.com/tuneinsight/lattigo/v3/ring"

	"github.com/SanthoshCheemala/FLARE/backend/utils"
)

// Adapter wraps the LE-PSI library for cleaner integration
type Adapter struct {
	maxWorkers int
}

func NewAdapter(maxWorkers int) *Adapter {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	return &Adapter{
		maxWorkers: maxWorkers,
	}
}

// ServerContext holds the PSI server state
type ServerContext struct {
	Hashes   []uint64
	TreePath string
	Ctx      *psi.ServerInitContext
	PP       *matrix.Vector
	Msg      *ring.Poly
	LE       *LE.LE
}

// ClientCiphertext represents encrypted client data
// We alias this to the library's type or wrap it
type ClientCiphertext = psi.Cxtx

// InitServer initializes the PSI server context with sanction data
func (a *Adapter) InitServer(ctx context.Context, sanctionSet []string, treePath string) (*ServerContext, error) {
	// Hash the sanction set
	hashes := HashDataPoints(sanctionSet)

	psiCtx, err := psi.ServerInitialize(hashes, treePath)
	if err != nil {
		return nil, fmt.Errorf("server initialize: %w", err)
	}
	pp, msg, le := psi.GetPublicParameters(psiCtx)

	serverCtx := &ServerContext{
		Hashes:   hashes,
		TreePath: treePath,
		Ctx:      psiCtx,
		PP:       pp,
		Msg:      msg,
		LE:       le,
	}

	return serverCtx, nil
}

// EncryptClient encrypts the client dataset with server's public parameters
func (a *Adapter) EncryptClient(ctx context.Context, clientSet []string, sc *ServerContext) ([]ClientCiphertext, error) {
	hashes := HashDataPoints(clientSet)

	ciphers := psi.ClientEncrypt(hashes, sc.PP, sc.Msg, sc.LE)

	return ciphers, nil
}

// DetectIntersection finds matching hashes between client and server sets
func (a *Adapter) DetectIntersection(ctx context.Context, sc *ServerContext, ciphertexts []ClientCiphertext) ([]uint64, error) {
	matches, err := psi.DetectIntersectionWithContext(sc.Ctx, ciphertexts)
	if err != nil {
		return nil, fmt.Errorf("detect intersection: %w", err)
	}

	return matches, nil
}

// EstimateMemory estimates memory requirements for PSI operation
func (a *Adapter) EstimateMemory(customerCount, sanctionCount int) float64 {
	// Rough estimate: ~35MB per 1000 records
	totalRecords := customerCount + sanctionCount
	return float64(totalRecords) * 35.0 / 1000.0
}

// GetWorkerCount returns the number of workers that will be used
func (a *Adapter) GetWorkerCount() int {
	return a.maxWorkers
}

// HashDataPoints converts strings to uint64 hashes using the utils package
func HashDataPoints(dataPoints []string) []uint64 {
	return utils.HashDataPoints(dataPoints)
}

// HashOne hashes a single string to uint64 using the utils package
func HashOne(data string) uint64 {
	hashes := utils.HashDataPoints([]string{data})
	if len(hashes) > 0 {
		return hashes[0]
	}
	return 0
}

// SerializeCustomer converts customer data to a normalized string for hashing
func SerializeCustomer(name, dob, country string) string {
	// Simple concatenation as requested
	return fmt.Sprintf("%s|%s|%s", normalizeString(name), dob, normalizeString(country))
}

// SerializeSanction converts sanction data to a normalized string for hashing
func SerializeSanction(name, dob, country, program string) string {
	return fmt.Sprintf("%s|%s|%s|%s", normalizeString(name), dob, normalizeString(country), normalizeString(program))
}

// SerializeDynamic creates a hash input string based on selected allowed columns.
// columns is a list of keys like "name", "dob", "country".
// values is a map where keys match the columns list.
func SerializeDynamic(values map[string]string, columns []string) string {
	var parts []string
	for _, col := range columns {
		val := values[col]
		if col == "name" || col == "country" || col == "program" {
			val = normalizeString(val)
		}
		parts = append(parts, val)
	}
	return strings.Join(parts, "|")
}

// normalizeString performs basic normalization (lowercase, trim)
func normalizeString(s string) string {
	// Normalize to lowercase and trim whitespace for consistent matching
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	return s
}

// ValidateMemoryRequirement checks if operation fits within memory constraints
func (a *Adapter) ValidateMemoryRequirement(customerCount, sanctionCount int, maxRAMGB float64) error {
	estimate := a.EstimateMemory(customerCount, sanctionCount)
	threshold := maxRAMGB * 0.85 // Use 85% threshold for safety

	if estimate > threshold {
		return fmt.Errorf("estimated memory %.2f GB exceeds threshold %.2f GB", estimate, threshold)
	}
	return nil
}

// SerializedServerParams aliases the library's serializable params
type SerializedServerParams = psi.SerializableParams

// SerializeParams serializes the server's public parameters using the library's method
func (a *Adapter) SerializeParams(sc *ServerContext) (*SerializedServerParams, error) {
	params := psi.SerializeParameters(sc.PP, sc.Msg, sc.LE)
	return params, nil
}

// DeserializeParams reconstructs the server parameters using the library's method
func (a *Adapter) DeserializeParams(params *SerializedServerParams) (*matrix.Vector, *ring.Poly, *LE.LE, error) {
	pp, msg, le, err := psi.DeserializeParameters(params)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("library deserialize failed: %w", err)
	}
	return pp, msg, le, nil
}

// PerformanceMonitor wraps the PSI library's performance monitor
type PerformanceMonitor struct {
	monitor *psi.PerformanceMonitor
}

// NewPerformanceMonitor creates a new performance monitor
func (a *Adapter) NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		monitor: psi.NewPerformanceMonitor(),
	}
}

// GetMetrics returns all performance metrics from the PSI library
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
	if pm.monitor == nil {
		return make(map[string]interface{})
	}
	return pm.monitor.GetMetrics()
}

// GetMemoryUsage returns current memory statistics
func (pm *PerformanceMonitor) GetMemoryUsage() map[string]interface{} {
	if pm.monitor == nil {
		return make(map[string]interface{})
	}
	return pm.monitor.GetMemoryUsage()
}

// GetThroughput calculates operations per second
func (pm *PerformanceMonitor) GetThroughput() float64 {
	if pm.monitor == nil {
		return 0.0
	}
	return pm.monitor.GetThroughput()
}

// ============================================================================
// BATCH PSI SUPPORT - For large datasets that exceed RAM limits
// ============================================================================

// BatchServerContext holds context for batch-processed PSI with large datasets
type BatchServerContext struct {
	Batches        []*ServerContext // Individual batch contexts
	BatchSize      int              // Records per batch
	TotalRecords   int              // Total server records
	TreePathPrefix string           // Prefix for batch tree files
}

// CalculateOptimalBatchSize determines batch size based on available RAM
func (a *Adapter) CalculateOptimalBatchSize() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get available system memory (in GB)
	// Use HeapSys as a proxy for available memory
	availableGB := float64(m.Sys) / 1024 / 1024 / 1024

	// Safety margin: use 75% of available RAM
	safeRAM_GB := availableGB * 0.75

	// Each server record needs ~32MB during initialization
	const RAM_PER_RECORD_GB = 0.032

	batchSize := int(safeRAM_GB / RAM_PER_RECORD_GB)

	// Bounds: minimum 50, maximum 1000
	if batchSize < 50 {
		batchSize = 50
	}
	if batchSize > 1000 {
		batchSize = 1000
	}

	return batchSize
}

// ShouldUseBatching determines if batch processing is needed
func (a *Adapter) ShouldUseBatching(recordCount int) bool {
	optimalBatch := a.CalculateOptimalBatchSize()
	return recordCount > optimalBatch
}

// InitServerBatched initializes PSI with batch processing for large datasets
// It automatically determines batch size based on available RAM
func (a *Adapter) InitServerBatched(ctx context.Context, sanctionSet []string, treePathPrefix string) (*BatchServerContext, error) {
	batchSize := a.CalculateOptimalBatchSize()
	totalRecords := len(sanctionSet)

	if totalRecords <= batchSize {
		// No batching needed, use single context
		sc, err := a.InitServer(ctx, sanctionSet, treePathPrefix+".db")
		if err != nil {
			return nil, err
		}
		return &BatchServerContext{
			Batches:        []*ServerContext{sc},
			BatchSize:      batchSize,
			TotalRecords:   totalRecords,
			TreePathPrefix: treePathPrefix,
		}, nil
	}

	// Calculate number of batches
	numBatches := (totalRecords + batchSize - 1) / batchSize

	bsc := &BatchServerContext{
		Batches:        make([]*ServerContext, 0, numBatches),
		BatchSize:      batchSize,
		TotalRecords:   totalRecords,
		TreePathPrefix: treePathPrefix,
	}

	// Initialize each batch sequentially to manage RAM
	for i := 0; i < numBatches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		batchData := sanctionSet[start:end]
		treePath := fmt.Sprintf("%s_batch%d.db", treePathPrefix, i)

		sc, err := a.InitServer(ctx, batchData, treePath)
		if err != nil {
			// Cleanup already initialized batches
			for _, prev := range bsc.Batches {
				if prev != nil && prev.TreePath != "" {
					// Cleanup would happen here
				}
			}
			return nil, fmt.Errorf("batch %d init failed: %w", i, err)
		}

		bsc.Batches = append(bsc.Batches, sc)

		// Force GC between batches to free memory
		runtime.GC()
	}

	return bsc, nil
}

// DetectIntersectionBatched runs intersection detection across all batches
// and aggregates the results
func (a *Adapter) DetectIntersectionBatched(ctx context.Context, bsc *BatchServerContext, clientSet []string) ([]uint64, error) {
	if len(bsc.Batches) == 1 {
		// Single batch, use standard detection
		ciphers, err := a.EncryptClient(ctx, clientSet, bsc.Batches[0])
		if err != nil {
			return nil, err
		}
		return a.DetectIntersection(ctx, bsc.Batches[0], ciphers)
	}

	// Multiple batches: aggregate matches
	allMatches := make(map[uint64]bool)

	for i, batch := range bsc.Batches {
		// Encrypt client data with this batch's parameters
		ciphers, err := a.EncryptClient(ctx, clientSet, batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d encryption failed: %w", i, err)
		}

		// Run intersection
		matches, err := a.DetectIntersection(ctx, batch, ciphers)
		if err != nil {
			return nil, fmt.Errorf("batch %d intersection failed: %w", i, err)
		}

		// Aggregate matches
		for _, m := range matches {
			allMatches[m] = true
		}

		// Force GC between batches
		runtime.GC()
	}

	// Convert map to slice
	result := make([]uint64, 0, len(allMatches))
	for hash := range allMatches {
		result = append(result, hash)
	}

	return result, nil
}

// CleanupBatchContext removes temporary files created during batch processing
func (a *Adapter) CleanupBatchContext(bsc *BatchServerContext) {
	if bsc == nil {
		return
	}
	// Note: Tree files are managed by the caller or cleaned up on server shutdown
	// This is a placeholder for any additional cleanup needed
}
