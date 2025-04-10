package LE

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)

// Decrypt decrypts a serialized ciphertext using the secret key and tree database
func Decrypt(leParams *LE, serializedCiphertext string, secretKey *matrix.Vector, treeDB *sql.DB, id uint64) (string, error) {
	// Parse the serialization format
	if !strings.HasPrefix(serializedCiphertext, "LE_ENC_") {
		return "", fmt.Errorf("invalid encryption format: %s", serializedCiphertext)
	}
	
	// In a real implementation, deserialize the ciphertext components
	// This is a placeholder implementation that just returns a dummy value
	
	// Generate witnesses for decryption
	witness1, witness2 := WitGen(treeDB, leParams, id)
	
	// Create dummy values for c0, c1, c, d
	c0 := make([]*matrix.Vector, leParams.Layers)
	c1 := make([]*matrix.Vector, leParams.Layers)
	for i := 0; i < leParams.Layers; i++ {
		c0[i] = matrix.NewVector(leParams.M, leParams.R)
		c1[i] = matrix.NewVector(leParams.M, leParams.R)
	}
	c := matrix.NewVector(leParams.N, leParams.R)
	d := leParams.R.NewPoly()
	
	// Decrypt the data
	decryptedPoly := Dec(leParams, secretKey, witness1, witness2, c0, c1, c, d)
	
	// Convert polynomial back to string
	return PolyToString(decryptedPoly, leParams.R), nil
}

// DeserializeEncryption deserializes encryption components from a string
func DeserializeEncryption(serialized string, r *ring.Ring) ([]*matrix.Vector, []*matrix.Vector, *matrix.Vector, *ring.Poly, error) {
	// In a real implementation, parse the serialized string and reconstruct the components
	// This is a placeholder implementation
	
	// Create dummy components
	c0 := make([]*matrix.Vector, 50) // Use appropriate layer count
	c1 := make([]*matrix.Vector, 50)
	for i := 0; i < 50; i++ {
		c0[i] = matrix.NewVector(4, r) // Use appropriate dimensions
		c1[i] = matrix.NewVector(4, r)
	}
	c := matrix.NewVector(4, r)
	d := r.NewPoly()
	
	return c0, c1, c, d, nil
}

// PolyToString converts a polynomial back to a string
func PolyToString(poly *ring.Poly, r *ring.Ring) string {
	// Simple decoding: each coefficient becomes a character
	// This is a simplified approach - real applications would use more sophisticated decoding
	var result strings.Builder
	
	for i := 0; i < r.N; i++ {
		coeff := poly.Coeffs[0][i]
		
		// Skip zero coefficients
		if coeff == 0 {
			continue
		}
		
		// Apply thresholding to determine if this should be a 0 or 1 bit
		// For laconic encryption, the coefficients are around q/2 for 1 and near 0 for 0
		if coeff > r.Modulus[0]/4 && coeff < 3*r.Modulus[0]/4 {
			// This is approximately a 1 bit
			c := rune(i % 128) // Map the position to an ASCII character
			if c >= 32 && c <= 126 { // Printable ASCII only
				result.WriteRune(c)
			}
		}
	}
	
	return result.String()
}
