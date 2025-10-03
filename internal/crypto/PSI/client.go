package psi

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"github.com/SanthoshCheemala/Crypto/hash"
	"github.com/SanthoshCheemala/FLARE/internal/storage"
	"github.com/SanthoshCheemala/FLARE/pkg/LE"
	"github.com/SanthoshCheemala/FLARE/pkg/matrix"
	_ "github.com/mattn/go-sqlite3"
)

// Client runs client side PSI (Production version - clean and efficient)
func Client(Client_Transaction []storage.Transaction, Server_Transaction []storage.Transaction, Treepath string) ([]storage.Transaction, error) {
	cSize := len(Client_Transaction)
	if cSize == 0 {
		return nil, errors.New("client transaction set is empty")
	}

	// setup LE params based on server size
	leParams, err := SetupLEParameters(len(Server_Transaction))
	if err != nil {
		return nil, fmt.Errorf("SetupLEParameters: %w", err)
	}

	db, err := sql.Open("sqlite3", Treepath)
	if err != nil {
		return nil, fmt.Errorf("open tree db: %w", err)
	}
	defer db.Close()

	if err := storage.InitializeTreeDB(db, leParams.Layers); err != nil {
		log.Printf("warning: InitializeTreeDB returned: %v\n", err)
	}

	publicKeys := make([]*matrix.Vector, cSize)
	privateKeys := make([]*matrix.Vector, cSize)
	hashedClient := make([]uint64, cSize)

	for i := 0; i < cSize; i++ {
		publicKeys[i], privateKeys[i] = leParams.KeyGen()

		// merge columns into string
		merge := ""
		sortedKeys := getSortedKeys(Client_Transaction[i].Data)
		for _, col := range sortedKeys {
			merge += Client_Transaction[i].Data[col]
		}

		// hash + mask into tree index
		H := hash.NewSHA256State()
		H.Sha256([]byte(merge))
		raw := binary.BigEndian.Uint64(H.Sum())

		var mask uint64
		bits := uint(leParams.Layers)
		if bits == 0 || bits >= 64 {
			mask = ^uint64(0)
		} else {
			mask = (uint64(1) << bits) - 1
		}
		hashedClient[i] = raw & mask

		// update DB
		LE.Upd(db, hashedClient[i], leParams.Layers, publicKeys[i], leParams)
	}

	// public parameters
	pp := LE.ReadFromDB(db, 0, 0, leParams).NTT(leParams.R)
	msg := matrix.NewRandomPolyBinary(leParams.R)

		// server ciphertexts
	ciphertexts := Server(pp, msg, Server_Transaction, leParams)

	// witnesses for client
	witnessesVec1 := make([][]*matrix.Vector, cSize)
	witnessesVec2 := make([][]*matrix.Vector, cSize)
	for i := 0; i < cSize; i++ {
		witnessesVec1[i], witnessesVec2[i] = LE.WitGen(db, leParams, hashedClient[i])
	}

	// intersection detection
	var Z []storage.Transaction
	intersectionMap := make(map[int]bool)

	for j := range ciphertexts {
		for k := 0; k < cSize; k++ {
			msg2 := LE.Dec(leParams, privateKeys[k], witnessesVec1[k], witnessesVec2[k],
				ciphertexts[j].C0, ciphertexts[j].C1, ciphertexts[j].C, ciphertexts[j].D)

			if CorrectnessCheck(msg2, msg, leParams) {
				if !intersectionMap[k] {
					Z = append(Z, Client_Transaction[k])
					intersectionMap[k] = true
				}
			}
		}
	}

	return Z, nil
}
