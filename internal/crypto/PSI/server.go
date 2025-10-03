package psi

import (
	"encoding/binary"
	"fmt"

	"github.com/SanthoshCheemala/Crypto/hash"
	"github.com/SanthoshCheemala/FLARE/internal/storage"
	"github.com/SanthoshCheemala/FLARE/pkg/LE"
	"github.com/SanthoshCheemala/FLARE/pkg/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)

// Server constructs encryptions for the server's set using the provided pp and msg.
func Server(pp *matrix.Vector, msg *ring.Poly, server_Transaction []storage.Transaction, le *LE.LE) []Cxtx {
	sSize := len(server_Transaction)
	mergedServer := make([]string, sSize)

	for idx, rec := range server_Transaction {
		merge := ""
		sortedKeys := getSortedKeys(rec.Data)
		for _, col := range sortedKeys {
			merge += rec.Data[col]
		}
		mergedServer[idx] = merge
	}

	// Hash server records to indices
	hashed := make([]uint64, sSize)
	for i := 0; i < sSize; i++ {
		H := hash.NewSHA256State()
		H.Sha256([]byte(mergedServer[i]))
		raw := binary.BigEndian.Uint64(H.Sum())

		var mask uint64
		bits := uint(le.Layers)
		if bits == 0 || bits >= 64 {
			mask = ^uint64(0)
		} else {
			mask = (uint64(1) << bits) - 1
		}
		hashed[i] = raw & mask
	}
	fmt.Printf("Server hashes: %v\n", hashed)

	// Build ciphertexts
	C := make([]Cxtx, sSize)
	for i := 0; i < sSize; i++ {
		r := make([]*matrix.Vector, le.Layers+1)
		for j := 0; j < le.Layers+1; j++ {
			r[j] = matrix.NewRandomVec(le.N, le.R, le.PRNG).NTT(le.R)
		}

		e := le.SamplerGaussian.ReadNew()
		e0 := make([]*matrix.Vector, le.Layers+1)
		e1 := make([]*matrix.Vector, le.Layers+1)
		for j := 0; j < le.Layers+1; j++ {
			if j == le.Layers {
				e0[j] = matrix.NewNoiseVec(le.M2, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
			} else {
				e0[j] = matrix.NewNoiseVec(le.M, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
			}
			e1[j] = matrix.NewNoiseVec(le.M, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
		}

		c0, c1, cvec, dpoly := LE.Enc(le, pp, hashed[i], msg, r, e0, e1, e)
		C[i] = Cxtx{C0: c0, C1: c1, C: cvec, D: dpoly}
	}

	return C
}

