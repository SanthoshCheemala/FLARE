package crypto

import (
	"github.com/tuneinsight/lattigo/v3/ring"
	"github.com/SanthoshCheemala/FLARE.git/matrix"
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



