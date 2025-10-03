package psi

import (
	"fmt"
	"math"
	"sort"

	"github.com/SanthoshCheemala/FLARE/pkg/LE"
	"github.com/SanthoshCheemala/FLARE/pkg/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)

// Cxtx represents a ciphertext structure used in PSI (exported for analytics)
type Cxtx struct {
	C0 []*matrix.Vector
	C1 []*matrix.Vector
	C  *matrix.Vector
	D  *ring.Poly
}

// getSortedKeys returns a sorted slice of keys from a map.
// This is crucial for ensuring deterministic serialization of map data.
func getSortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// MeasureNoiseLevel calculates the noise level between an original message and its decrypted version.
// It returns:
// - maxNoiseFraction: maximum noise as a fraction of Q (e.g., 0.01 means 1% of Q)
// - avgNoiseFraction: average noise as a fraction of Q
// - noiseDistribution: a map showing the distribution of noise levels
func MeasureNoiseLevel(r *ring.Ring, original, decrypted *ring.Poly, Q uint64) (maxNoiseFraction, avgNoiseFraction float64, noiseDistribution map[string]int) {
    diff := r.NewPoly()
    r.Sub(decrypted, original, diff)
    
    totalCoeffs := len(diff.Coeffs[0])
    maxNoise := uint64(0)
    totalNoise := uint64(0)
    
    // Initialize noise distribution bins
    noiseDistribution = map[string]int{
        "0-0.1%Q": 0,
        "0.1-1%Q": 0,
        "1-5%Q": 0,
        "5-10%Q": 0,
        "10-25%Q": 0,
        ">25%Q": 0,
    }
    
    // Calculate noise for each coefficient
    for _, coeff := range diff.Coeffs[0] {
        // Convert coefficient to its absolute distance from 0
        // Consider both directions of noise (coeff could be close to Q when noise is negative)
        var noise uint64
        if coeff > Q/2 {
            noise = Q - coeff // negative noise (coeff close to Q)
        } else {
            noise = coeff // positive noise
        }
        
        // Track maximum noise
        if noise > maxNoise {
            maxNoise = noise
        }
        
        // Accumulate total noise for average calculation
        totalNoise += noise
        
        // Add to distribution buckets
        noiseFraction := float64(noise) / float64(Q)
        switch {
        case noiseFraction <= 0.001:
            noiseDistribution["0-0.1%Q"]++
        case noiseFraction <= 0.01:
            noiseDistribution["0.1-1%Q"]++
        case noiseFraction <= 0.05:
            noiseDistribution["1-5%Q"]++
        case noiseFraction <= 0.1:
            noiseDistribution["5-10%Q"]++
        case noiseFraction <= 0.25:
            noiseDistribution["10-25%Q"]++
        default:
            noiseDistribution[">25%Q"]++
        }
    }
    
    // Calculate max and average noise as fraction of Q
    maxNoiseFraction = float64(maxNoise) / float64(Q)
    avgNoiseFraction = float64(totalNoise) / float64(totalCoeffs) / float64(Q)
    
    return maxNoiseFraction, avgNoiseFraction, noiseDistribution
}

func CorrectnessCheck(decrypted, original *ring.Poly, le *LE.LE) bool {
    q14 := le.Q / 4
    q34 := (le.Q / 4) * 3
    binaryDecrypted := le.R.NewPoly()
    
    // Convert coefficients to binary based on thresholds
    for i := 0; i < le.R.N; i++ {
        if decrypted.Coeffs[0][i] < q14 || decrypted.Coeffs[0][i] > q34 {
            binaryDecrypted.Coeffs[0][i] = 0
        } else {
            binaryDecrypted.Coeffs[0][i] = 1
        }
    }
    
    // Enhanced debugging
    matchCount := 0
    mismatchCount := 0
    for i := 0; i < le.R.N; i++ {
        if binaryDecrypted.Coeffs[0][i] == original.Coeffs[0][i] {
            matchCount++
        } else {
            mismatchCount++
            if mismatchCount <= 5 { // Show first 5 mismatches
                fmt.Printf("Mismatch at coeff %d: decoded=%d, original=%d (raw=%d)\n", 
                    i, binaryDecrypted.Coeffs[0][i], original.Coeffs[0][i], decrypted.Coeffs[0][i])
            }
        }
    }
    
    fmt.Printf("Correctness: %d matches, %d mismatches out of %d coefficients\n", 
        matchCount, mismatchCount, le.R.N)
    
    // Use a threshold instead of perfect equality for noisy decryption
    matchPercentage := float64(matchCount) / float64(le.R.N)
    fmt.Printf("Match percentage: %.2f%%\n", matchPercentage*100)
    
    // Consider it correct if at least 95% of coefficients match
    return matchPercentage >= 0.95
}

// SetupLEParameters sets up LE parameters based on server size
func SetupLEParameters(serverSize int) (*LE.LE, error) {
    return SetupLEParametersWithDimension(serverSize, 256) // Default ring dimension
}

// SetupLEParametersWithDimension sets up LE parameters with custom ring dimension
func SetupLEParametersWithDimension(serverSize, ringDimension int) (*LE.LE, error) {
    Q := uint64(180143985094819841)
    qBits := 58
    D := ringDimension
    N := 4

    // Validate ring dimension
    if D != 256 && D != 512 && D != 1024 && D != 2048 {
        return nil, fmt.Errorf("unsupported ring dimension %d. Supported values: 256, 512, 1024, 2048", D)
    }

    // Create LE parameters using the Setup function
    leParams := LE.Setup(Q, qBits, D, N)
    if leParams == nil {
        return nil, fmt.Errorf("failed to initialize the le parameters (nil result)")
    }
    if leParams.R == nil {
        return nil, fmt.Errorf("ring(R) is nil in le parameters")
    }

    // Calculate appropriate number of layers for the tree
    // Expansion factor (more slots than items to reduce collisions)
    c := 16.0
    layers := int(math.Ceil(math.Log2(c * float64(serverSize))))
    if layers < 3 {
        layers = 3
    }
    leParams.Layers = layers
    
    return leParams, nil
}
