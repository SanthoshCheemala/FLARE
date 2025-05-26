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


func EncryptTransactions(transactions []storage.Transaction,columns []string,TreeDbPath, SercretPath string)([]storage.Transaction,error){
	leParams,err := SetupLEParameters()
	if err != nil{
		log.Fatal(err)
	}
	fmt.Printf("Generating Encryption Key Pairs...")
	publicKey, secretKey := leParams.KeyGen()
	if publicKey == nil || len(publicKey.Elements) == 0{
		return nil,fmt.Errorf("failed to generate valid key Pairs")
	}
	fmt.Printf("Saving secret key to %s",SercretPath)
	if err := storage.SaveSecretkey(secretKey,SercretPath); err != nil{
		return nil,fmt.Errorf("Failed to save secret key: %w",err)
	}
	encryptTransactions := make([]storage.Transaction,len(transactions))

	treeDb,err := sql.Open("sqlite3",TreeDbPath)

	if err !=  nil{
		fmt.Println("we couldn't open treeDb try again")	
	} else {
		defer treeDb.Close()
		storage.InitializeTreeDB(treeDb,leParams.Layers)
	}

	successCount := 0
	errorCount := 0

	for i,trans := range transactions{
		if i > 0 && i%10 == 0{
			fmt.Printf("Encrypted %d/%d Transactions \n",i,len(transactions))
		}
		encryptedTrans := storage.Transaction{
			Data: make(map[string]string),
		}

		for _,col := range columns{
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
				fmt.Printf("Encryption Failed for %s: %v\n",col,EncErr)
				EncryptedStr = fmt.Sprintf("PLAIN_%s",dataStr)
				errorCount++;
			} else {
				dBytes,err := d.MarshalBinary()
				if err != nil{
					fmt.Printf("Serilization Failed for %s: %v\n",col,err)
					EncryptedStr = fmt.Sprintf("PLAIN_%s",dataStr)
					errorCount++;
				} else {
					EncryptedStr = SerilizeEncryption(dBytes)
					successCount++;
				}
			}
			encryptedTrans.Data[col] = EncryptedStr

		}
		encryptTransactions[i] = encryptedTrans
	}
		fmt.Printf("Performed Encrypted Transactions with successfull encryptions: %d, Errors: %d",successCount,errorCount)
		fmt.Println("All transactions are Proccessed")
		return encryptTransactions,nil
}