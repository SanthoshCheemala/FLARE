package crypto

import (
	"database/sql"
	"encoding/binary"
	"fmt"

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

	publikeys := make([]*matrix.Vector,c_size)
	privateKeys := make([]*matrix.Vector,c_size)
	Hashed_Client_Transactions := make([]uint64,c_size)
	merged_Client_Transaction := make([]string,c_size)

	for i := 0; i < c_size; i++ {
		publikeys[i],privateKeys[i] = leParams.KeyGen()
		merge := ""
		sortedKeys := getSortedKeys(Client_Transaction[i].Data)
		for _, col := range sortedKeys {
			merge += Client_Transaction[i].Data[col]
		}
		merged_Client_Transaction[i] = merge
		H := hash.NewSHA256State()
		H.Sha256([]byte(merged_Client_Transaction[i]))
		Hashed_Client_Transactions[i] = binary.BigEndian.Uint64(H.Sum())
		LE.Upd(db,Hashed_Client_Transactions[i],leParams.Layers,publikeys[i],leParams)
	}
	fmt.Println(Hashed_Client_Transactions)
	pp := LE.ReadFromDB(db,0,0,leParams).NTT(leParams.R)
	msg := matrix.NewRandomPolyBinary(leParams.R)
	
	Cxtx := Laconic_PSI_server(pp,msg,Server_Transaction,leParams)
    witnesses_vec1 := make([][]*matrix.Vector, c_size)
	witnesses_vec2 := make([][]*matrix.Vector, c_size)
	for i := 0; i < c_size; i++ {
		witnesses_vec1[i],witnesses_vec2[i] = LE.WitGen(db,leParams,Hashed_Client_Transactions[i])
	}
	
    Z := make([]storage.Transaction, 0)
    intersection_map := make(map[int]bool)
    
    var totalMaxNoise, totalAvgNoise float64
    var totalMatches int

    for j := range Cxtx { 
        for k := 0; k < c_size; k++ { 
            msg2 := LE.Dec(leParams, privateKeys[k], witnesses_vec1[k], witnesses_vec2[k], Cxtx[j].c0, Cxtx[j].c1, Cxtx[j].c, Cxtx[j].d)

            maxNoise, avgNoise, noiseDistribution := MeasureNoiseLevel(leParams.R, msg, msg2, leParams.Q)

                    totalMaxNoise += maxNoise
                    totalAvgNoise += avgNoise
                    totalMatches++
                    
                    fmt.Printf("Match found with max noise: %.6f%% of Q, avg noise: %.6f%% of Q\n", 
                            maxNoise*100, avgNoise*100)
                    PrettyPrintNoiseDistribution(noiseDistribution,len(msg.Coeffs[0]))
                
            if CorrectnessCheck( msg2, msg, leParams) {
                if _, ok := intersection_map[k]; !ok {
                    Z = append(Z, Client_Transaction[k])
                    intersection_map[k] = true
                }
            }
        }
    }

    // if totalMatches > 0 {
    //     fmt.Printf("Overall noise statistics across %d matches:\n", totalMatches)
    //     fmt.Printf("  - Average max noise: %.6f%% of Q\n", (totalMaxNoise/float64(totalMatches))*100)
    //     fmt.Printf("  - Average mean noise: %.6f%% of Q\n", (totalAvgNoise/float64(totalMatches))*100)
    // }
    
    return Z, nil
}

func Laconic_PSI_server(pp *matrix.Vector,msg *ring.Poly,server_Transaction []storage.Transaction,le *LE.LE) ([]cxtx) {
    s_size := len(server_Transaction)
    merged_Server_Transaction := make([]string,s_size)
        for index,record := range server_Transaction {
            merge := ""
            sortedKeys := getSortedKeys(record.Data)
            for _, col := range sortedKeys {
                merge += record.Data[col]
            }
            merged_Server_Transaction[index] = merge
        }
    Hashed_Transactions := make([]uint64,s_size)
    for i := 0; i < s_size; i++ {
        H := hash.NewSHA256State()
        H.Sha256([]byte(merged_Server_Transaction[i]))
        Hashed_Transactions[i] = binary.BigEndian.Uint64(H.Sum())
    }
    fmt.Println(Hashed_Transactions)

    // msgNTT := msg.CopyNew()
    // le.R.NTT(msgNTT, msgNTT)
    
    Cxtx := make([]cxtx,s_size)
    for i := 0; i < s_size; i++ {
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
    // fmt.Println(Cxtx[0].c0,Cxtx[0].c1)
    return Cxtx
}