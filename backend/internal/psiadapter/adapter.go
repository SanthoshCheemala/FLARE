package psiadapter

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"runtime"
	"strings"

	"github.com/SanthoshCheemala/LE-PSI/pkg/LE"
	"github.com/SanthoshCheemala/LE-PSI/pkg/matrix"
	"github.com/SanthoshCheemala/LE-PSI/pkg/psi"
	"github.com/tuneinsight/lattigo/v3/ring"

	"github.com/SanthoshCheemala/FLARE/backend/internal/models"
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

// HashDataPoints converts strings to uint64 hashes
func HashDataPoints(dataPoints []string) []uint64 {
	hashes := make([]uint64, len(dataPoints))
	for i, dp := range dataPoints {
		hashes[i] = HashOne(dp)
	}
	return hashes
}

// HashOne hashes a single string to uint64
func HashOne(data string) uint64 {
	h := sha256.Sum256([]byte(data))
	return binary.BigEndian.Uint64(h[:8])
}

// SerializeCustomer converts customer data to a normalized string for hashing
func SerializeCustomer(name, dob, country string) string {
	dp := models.PSIDataPoint{
		Name:    normalizeString(name),
		DOB:     dob,
		Country: normalizeString(country),
	}
	s, _ := utils.SerializeData(dp)
	return s
}

// SerializeSanction converts sanction data to a normalized string for hashing
func SerializeSanction(name, dob, country, program string) string {
	// Use same format as customer for matching - ignore program field
	dp := models.PSIDataPoint{
		Name:    normalizeString(name),
		DOB:     dob,
		Country: normalizeString(country),
	}
	s, _ := utils.SerializeData(dp)
	return s
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
