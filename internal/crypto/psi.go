package crypto

import (
	"database/sql"
	"encoding/binary"

	"github.com/SanthoshCheemala/Crypto/hash"
	"github.com/SanthoshCheemala/FLARE.git/internal/storage"
	"github.com/SanthoshCheemala/FLARE.git/pkg/LE"
	"github.com/SanthoshCheemala/FLARE.git/pkg/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)
type cxtx struct{
	c0 []*matrix.Vector
	c1 []*matrix.Vector
	c *matrix.Vector
	d *ring.Poly
}

func Laconic_PSI(Client_Transaction []storage.Transaction,Server_Transaction []storage.Transaction, Treepath string) (intersection_trans []storage.Transaction,err error) {
	c_size := len(Client_Transaction)
	leParams,err := SetupLEParameters(c_size)
	if err != nil {
		return nil,err
	}
	db, _ := sql.Open("sqlite3", Treepath)
	storage.InitializeTreeDB(db,leParams.Layers)

	if err != nil {
		return nil, err
	}

	publikeys := make([]*matrix.Vector,c_size)
	privateKeys := make([]*matrix.Vector,c_size)

	for i := 0; i < c_size; i++ {
		publikeys[i],privateKeys[i] = leParams.KeyGen()
	}
	merged_Client_Transaction := make([]string,c_size)
		for index,record := range Client_Transaction {
			merge := ""
			for col := range record.Data{
				merge += record.Data[col]
			}
			merged_Client_Transaction[index] = merge
		}
	Hashed_Transactions := make([]uint64,c_size)
	
	for i := 0; i < c_size; i++ {
		H := hash.NewSHA256State()
		H.Sha256([]byte(merged_Client_Transaction[i]))
		
		Hashed_Transactions[i] = binary.BigEndian.Uint64(H.Sum())
	}

	LE.Upd(db,Hashed_Transactions[0],leParams.Layers,publikeys[0],leParams)
	pp := LE.ReadFromDB(db,0,0,leParams).NTT(leParams.R)
	msg := matrix.NewRandomPolyBinary(leParams.R)
	Cxtx := Laconic_PSI_server(pp,msg,Server_Transaction,leParams)
	// PSI_SET := []int{}
	return nil,nil
}

func Laconic_PSI_server(pp *matrix.Vector,msg *ring.Poly,server_Transaction []storage.Transaction,le *LE.LE) ([]cxtx) {
	s_size := len(server_Transaction)
	merged_Client_Transaction := make([]string,s_size)
		for index,record := range server_Transaction {
			merge := ""
			for col := range record.Data{
				merge += record.Data[col]
			}
			merged_Client_Transaction[index] = merge
		}
	Hashed_Transactions := make([]uint64,s_size)
	
	for i := 0; i <= s_size; i++ {
		H := hash.NewSHA256State()
		H.Sha256([]byte(merged_Client_Transaction[i]))
		Hashed_Transactions[i] = binary.BigEndian.Uint64(H.Sum())
	}
	Cxtx := make([]cxtx,s_size)
	for i := 0; i <= s_size; i++ {
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
		Cxtx[i].c0,Cxtx[i].c1,Cxtx[i].c,Cxtx[i].d = LE.Enc(le,pp,Hashed_Transactions[i],msg,r,e0,e1,e)
	}
	return Cxtx;
}