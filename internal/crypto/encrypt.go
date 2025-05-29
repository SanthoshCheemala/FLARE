package crypto

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/SanthoshCheemala/FLARE.git/internal/storage"
	"github.com/tuneinsight/lattigo/v3/ring"
)



func EncryptTransactions(transactions []storage.Transaction,columns []string,TreeDbPath, SercretPath string)([]storage.Transaction,[]storage.MergedTransaction,error){
	leParams,err := SetupLEParameters()
	if err != nil{
		log.Fatal(err)
	}
	publicKey, secretKey := leParams.KeyGen()
	if publicKey == nil || len(publicKey.Elements) == 0{
		return nil,nil,fmt.Errorf("failed to generate valid key Pairs")
	}
	if err := storage.SaveSecretkey(secretKey,SercretPath); err != nil{
		return nil,nil,fmt.Errorf("failed to save secret key: %w",err)
	}
	encryptTransactions := make([]storage.Transaction,len(transactions))
	encryptionMergedTrans := make([]storage.MergedTransaction,len(transactions))
	treeDb,err := sql.Open("sqlite3",TreeDbPath)

	if err ==  nil{
		defer treeDb.Close()
		storage.InitializeTreeDB(treeDb,leParams.Layers)
	}

	successCount := 0
	errorCount := 0

	for i,trans := range transactions{
		mergedEncryptTran := ""
		encryptedTrans := storage.Transaction{
			Data: make(map[string]string),
		}

		for _,col := range columns{
			mergedEncryptTran += trans.Data[col]
			dataStr := trans.Data[col]
			dataPloy := StringToPoly(dataStr,leParams.R)

			var d *ring.Poly
			var EncErr error

			func(){
				defer func(){
					if r := recover(); r != nil{
						EncErr = fmt.Errorf("panic in direct Encryption: %v",r)
					}
				}()

				d = leParams.R.NewPoly()
				nonce := make([]byte,8)
				if _,err := rand.Read(nonce); err == nil{
					nonceStr := base64.StdEncoding.EncodeToString(nonce)
					salt := fmt.Sprintf("%s-%d-%s-%s",col,i,nonceStr,dataStr)
					saltPoly := StringToPoly(salt,leParams.R)

					leParams.R.Add(dataPloy,publicKey.Elements[0],d)
					leParams.R.Add(d,saltPoly,d)
					leParams.R.NTT(d,d)
				}
				if len(publicKey.Elements) > 1{
					temp := leParams.R.NewPoly()
					leParams.R.MulCoeffs(d,publicKey.Elements[1],temp)
					d = temp
				}
			}()
			var EncryptedStr string
			if EncErr != nil{
				EncryptedStr = fmt.Sprintf("PLAIN_%s",dataStr)
				errorCount++;
			} else {
				dBytes,err := d.MarshalBinary()
				if err != nil{
					EncryptedStr = fmt.Sprintf("PLAIN_%s",dataStr)
					errorCount++;
				} else {
					EncryptedStr = SerilizeEncryption(dBytes)
					successCount++;
				}
			}

			encryptedTrans.Data[col] = EncryptedStr

		}

		mergedDataPoly := StringToPoly(mergedEncryptTran,leParams.R)
		var d2 *ring.Poly
		var EncErr2 error

		func(){
			defer func(){
				if r := recover(); r != nil{
					EncErr2 = fmt.Errorf("panic in direct Encryption: %v",r)
				}
			}()

			d2 = leParams.R.NewPoly()
			nonce := make([]byte,8)
			if _,err := rand.Read(nonce); err == nil{
				nonceStr := base64.StdEncoding.EncodeToString(nonce)
				salt := fmt.Sprintf("%d-%s",i,nonceStr)
				saltPoly := StringToPoly(salt,leParams.R)

				leParams.R.Add(mergedDataPoly,publicKey.Elements[0],d2)
				leParams.R.Add(d2,saltPoly,d2)
				leParams.R.NTT(d2,d2)
			}
			if len(publicKey.Elements) > 1{
				temp := leParams.R.NewPoly()
				leParams.R.MulCoeffs(d2,publicKey.Elements[1],temp)
				d2 = temp
			}
		}()
		var EncryptedStr2 string
		if EncErr2 != nil{
			EncryptedStr2 = fmt.Sprintf("PLAIN_%s",mergedEncryptTran)
			errorCount++;
		} else {
			dBytes,err := d2.MarshalBinary()
			if err != nil{
				EncryptedStr2 = fmt.Sprintf("PLAIN_%s",mergedEncryptTran)
				errorCount++;
			} else {
				EncryptedStr2 = SerilizeEncryption(dBytes)
				successCount++;
			}
		}
		encryptionMergedTrans[i].Data = EncryptedStr2
		encryptionMergedTrans[i].Index = i
		encryptTransactions[i] = encryptedTrans
	}
		fmt.Printf("Performed Encrypted Transactions with successfull encryptions: %d, Errors: %d",successCount,errorCount)
		fmt.Println("All transactions are Proccessed")
		return encryptTransactions,encryptionMergedTrans,nil
}