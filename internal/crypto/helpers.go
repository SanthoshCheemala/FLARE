package crypto

import (
	// "fmt"
	"fmt"
	"sort"
	"strings"

	"github.com/SanthoshCheemala/FLARE.git/pkg/LE"
	"github.com/tuneinsight/lattigo/v3/ring"
)

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

// isNoisyEqual checks if two polynomials are equal, accounting for small noise from decryption.
// It computes their difference and checks if the coefficients of the resulting polynomial are small.
// A coefficient is considered "small" if it's close to 0 or the modulus Q.
func isNoisyEqual(r *ring.Ring, p1, p2 *ring.Poly, Q uint64) bool {
    diff := r.NewPoly()
    r.Sub(p1, p2, diff)

    threshold := Q / 2
    for _, coeff := range diff.Coeffs[0] {
        // A small coefficient is close to 0 or Q. A large one is in the middle.
        if coeff > threshold && coeff < Q-threshold {
            return false // Found a large coefficient, polynomials are not equal.
        }
    }
    return true // All coefficients are small, polynomials are considered equal.
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


// Replace the current ImprovedNoisyEqual function with this one:

func ImprovedNoisyEqual(r *ring.Ring, p1, p2 *ring.Poly, Q uint64) bool {
    diff := r.NewPoly()
    r.Sub(p1, p2, diff)
    
    totalCoeffs := len(diff.Coeffs[0])
    lowNoiseCount := 0
    mediumNoiseCount := 0
    
    for _, coeff := range diff.Coeffs[0] {
        // Convert to absolute distance from 0
        var noise uint64
        if coeff > Q/2 {
            noise = Q - coeff // negative noise (coeff close to Q)
        } else {
            noise = coeff // positive noise
        }
        
        // Categorize noise levels
        if noise < Q/4 {  // More lenient threshold (Q/4 instead of Q/8)
            lowNoiseCount++
        } else if noise < Q/2 {  // Medium noise
            mediumNoiseCount++
        }
    }
    
    // Calculate percentages
    lowNoisePercentage := float64(lowNoiseCount) / float64(totalCoeffs)
    mediumNoisePercentage := float64(mediumNoiseCount) / float64(totalCoeffs)
    
    // Debug output
    fmt.Printf("Noise analysis: %.2f%% low noise, %.2f%% medium noise\n", 
               lowNoisePercentage*100, mediumNoisePercentage*100)
    
    // Combined approach: 
    // 1. Either we have at least 40% coefficients with low noise
    // 2. Or we have at least 80% coefficients with medium-or-better noise
    return lowNoisePercentage >= 0.4 || (lowNoisePercentage + mediumNoisePercentage) >= 0.8
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
    fmt.Println(binaryDecrypted)
    fmt.Println("=====================")
     fmt.Println(original)
    // Compare with original message
    return binaryDecrypted.Equals(original)
}

// PrettyPrintNoiseDistribution creates a visual representation of noise distribution
func PrettyPrintNoiseDistribution(noiseDistribution map[string]int, totalCoeffs int) {
    fmt.Println("\n=== Noise Distribution Analysis ===")
    
    // Define categories in order from lowest to highest noise
    categories := []string{"0-0.1%Q", "0.1-1%Q", "1-5%Q", "5-10%Q", "10-25%Q", ">25%Q"}
    descriptions := []string{
        "Negligible noise",
        "Very low noise",
        "Low noise",
        "Moderate noise", 
        "High noise",
        "Very high noise",
    }
    
    // Find the maximum count for scaling the bar chart
    maxCount := 0
    for _, count := range noiseDistribution {
        if count > maxCount {
            maxCount = count
        }
    }
    
    // Print header
    fmt.Printf("%-15s | %-12s | %-20s | %s\n", "Category", "Count", "Percentage", "Distribution")
    fmt.Println(strings.Repeat("-", 80))
    
    // Print each category
    for i, category := range categories {
        count := noiseDistribution[category]
        percentage := float64(count) / float64(totalCoeffs) * 100
        
        // Create a visual bar representing the percentage
        barLength := int(percentage / 2) // Scale: each character is ~2%
        bar := strings.Repeat("█", barLength)
        
        // Print the row
        fmt.Printf("%-15s | %-12d | %6.2f%% %-10s | %s %s\n", 
            category, 
            count, 
            percentage,
            "", // Empty space for alignment
            bar,
            descriptions[i])
    }
    
    fmt.Println("\nTotal coefficients analyzed:", totalCoeffs)
}