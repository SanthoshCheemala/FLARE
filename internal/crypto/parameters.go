package crypto

import (
	"fmt"
	"math"

	"github.com/SanthoshCheemala/FLARE/pkg/LE"
)

// SetupLEParameters sets up Laconic Encryption parameters based on expected set size.
func SetupLEParameters(size int) (*LE.LE, error) {
	Q := uint64(180143985094819841)
	qBits := 58
	D := 256
	N := 4

	var leParams *LE.LE
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic in LE.setup: %v", r)
				fmt.Printf("Recovered from Panic in LE.setup: %v\n", r)
			}
		}()
		fmt.Println("Setting up LE with Parameters Q =", Q, "qBits =", qBits, "D =", D, "N =", N)
		leParams = LE.Setup(Q, qBits, D, N)
	}()
	if err != nil {
		return nil, err
	}
	if leParams == nil {
		return nil, fmt.Errorf("failed to initialize the le parameters (nil result)")
	}
	if leParams.R == nil {
		return nil, fmt.Errorf("ring(R) is nil in le parameters")
	}

	// Expansion factor (more slots than items to reduce collisions)
	c := 16.0
	// Compute layers: smallest power of two >= c * size
	leParams.Layers = int(math.Ceil(math.Log2(c * float64(size))))

	// Derived values
	numSlots := 1 << leParams.Layers
	loadFactor := float64(size) / float64(numSlots)

	// Approximate collision probability (balls-into-bins model)
	// P(no collision) ≈ exp(-size^2 / (2 * numSlots))
	// So collisionProb ≈ 1 - exp(-m^2 / (2N))
	m := float64(size)
	Nf := float64(numSlots)
	collisionProb := 1.0 - math.Exp(-(m*m)/(2*Nf))

	// Print results
	fmt.Println("Successfully initialized the LE parameters:")
	fmt.Printf(" - Ring Dimension: %d\n", D)
	fmt.Printf(" - Modulus Q: %d\n", Q)
	fmt.Printf(" - Matrix Dimension N: %d\n", N)
	fmt.Printf(" - qBits: %d\n", qBits)
	fmt.Printf(" - Layers: %d (slots = %d)\n", leParams.Layers, numSlots)
	fmt.Printf(" - Load Factor: %.6f (items/slot)\n", loadFactor)
	fmt.Printf(" - Estimated Collision Probability: %.6e\n", collisionProb)

	return leParams, nil
}
