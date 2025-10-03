package crypto

import (
	psi "github.com/SanthoshCheemala/FLARE/internal/crypto/PSI"
	"github.com/SanthoshCheemala/FLARE/internal/storage"
)

// Laconic_PSI runs client side PSI (Production version - clean and efficient)
// This is a wrapper function that calls the PSI client implementation
func Laconic_PSI(Client_Transaction []storage.Transaction, Server_Transaction []storage.Transaction, Treepath string) ([]storage.Transaction, error) {
	return psi.Client(Client_Transaction, Server_Transaction, Treepath)
}
