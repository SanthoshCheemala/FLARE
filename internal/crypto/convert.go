package crypto

import (
	"encoding/base64"
	"fmt"

	"github.com/tuneinsight/lattigo/v3/ring"
	// "github.com/SanthoshCheemala/FLARE/pkg/matrix"
)

func StringToPoly(s string,r *ring.Ring) *ring.Poly{
	poly := r.NewPoly()

	for i,c := range s{
		if i < r.N{
			poly.Coeffs[0][i] = uint64(c) % r.Modulus[0]
		}
	}
	return poly
}

func SerilizeEncryption(dBytes []byte) string{
	checkSum := uint32(0)
	for _,b := range dBytes{
		checkSum += uint32(b)
	}
	encodedData := base64.StdEncoding.EncodeToString(dBytes)
	encryptedStr := fmt.Sprintf("LE_ENCv1_%d_%s",checkSum,encodedData)
	return encryptedStr
}



