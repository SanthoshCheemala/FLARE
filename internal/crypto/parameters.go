package crypto

import (
	"fmt"

	"github.com/SanthoshCheemala/FLARE.git/pkg/le"
)

/* SetupLEParameters creates lattice encryption parameters with safer values
   Uses values that have been tested and confirmed to work with the current implementation
   Includes panic recovery to prevent crashes during initialization */
func SetupLEParameters() (*le.LE, error) {
	Q := uint64(65537)    // A small prime modulus
	qBits := 8            // Small bit size that works reliably
	D := 32               // Ring dimension that avoids index errors
	N := 2                // Matrix dimension that works with our approach
	
	var leParams *le.LE
	var err error
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic in LE.Setup: %v", r)
				fmt.Printf("Recovered from panic in LE.Setup: %v\n", r)
			}
		}()
		
		fmt.Println("Setting up LE with parameters: Q=", Q, "qBits=", qBits, "D=", D, "N=", N)
		leParams = le.Setup(Q, qBits, D, N)
	}()
	
	if err != nil {
		return nil, err
	}
	
	if leParams == nil {
		return nil, fmt.Errorf("failed to initialize LE parameters (nil result)")
	}
	
	if leParams.R == nil {
		return nil, fmt.Errorf("ring (R) is nil in LE parameters")
	}
	
	leParams.Layers = 2   // Minimal tree depth for efficiency
	
	fmt.Println("Successfully initialized LE parameters:")
	fmt.Printf("  - Ring dimension: %d\n", D)
	fmt.Printf("  - Modulus Q: %d\n", Q)
	fmt.Printf("  - Matrix dimension N: %d\n", N)
	fmt.Printf("  - QBits: %d\n", qBits)
	fmt.Printf("  - Layers: %d\n", leParams.Layers)
	
	return leParams, nil
}
