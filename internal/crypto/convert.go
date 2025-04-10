package crypto

import (
	"encoding/base64"
	"fmt"

	"github.com/SanthoshCheemala/FLARE.git/pkg/le/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)

/* StringToPoly converts a string to a polynomial for encryption
   Maps each character to a coefficient in the polynomial */
func StringToPoly(s string, r *ring.Ring) *ring.Poly {
	poly := r.NewPoly()
	
	for i, c := range s {
		if i < r.N {
			poly.Coeffs[0][i] = uint64(c) % r.Modulus[0]
		}
	}
	return poly
}

/* SerializeEncryption creates a string representation of encryption components
   Handles polynomial serialization, checksum calculation, and base64 encoding */
func SerializeEncryption(c0, c1 []*matrix.Vector, c *matrix.Vector, d *ring.Poly) (string, error) {
	dBytes, err := d.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to serialize d polynomial: %w", err)
	}
	
	checksum := uint32(0)
	for _, b := range dBytes {
		checksum += uint32(b)
	}
	
	encodedData := base64.StdEncoding.EncodeToString(dBytes)
	
	return fmt.Sprintf("LE_ENCv1_%d_%s", checksum, encodedData), nil
}
