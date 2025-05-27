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


func EncryptTransactions(transactions []storage.Transaction,columns []string,TreeDbPath, SercretPath string)([]storage.Transaction,[]string,error){
	leParams,err := SetupLEParameters()
	if err != nil{
		log.Fatal(err)
	}
	fmt.Printf("Generating Encryption Key Pairs...")
	publicKey, secretKey := leParams.KeyGen()
	if publicKey == nil || len(publicKey.Elements) == 0{
		return nil,nil,fmt.Errorf("failed to generate valid key Pairs")
	}
	fmt.Printf("Saving secret key to %s",SercretPath)
	if err := storage.SaveSecretkey(secretKey,SercretPath); err != nil{
		return nil,nil,fmt.Errorf("failed to save secret key: %w",err)
	}
	encryptTransactions := make([]storage.Transaction,len(transactions))
	encryptTransactions2 := make([]string,len(transactions))
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
		mergedEncryptedTransaction := ""
		if i > 0 && i%10 == 0{
			fmt.Printf("Encrypted %d/%d Transactions \n",i,len(transactions))
		}
		encryptedTrans := storage.Transaction{
			Data: make(map[string]string),
		}

		for _,col := range columns{
			mergedEncryptedTransaction += trans.Data[col]
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
		mergedDataPoly := StringToPoly(mergedEncryptedTransaction,leParams.R)
		fmt.Println("merged Transaction string size",len(mergedEncryptedTransaction))
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
				salt := fmt.Sprintf("%d-%s-%s",i,nonceStr)
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
				fmt.Printf("Encryption Failed for: %v\n",EncErr2)
				EncryptedStr2 = fmt.Sprintf("PLAIN_%s",mergedEncryptedTransaction)
				errorCount++;
			} else {
				dBytes,err := d2.MarshalBinary()
				if err != nil{
					fmt.Printf("Serilization Failed : %v\n",err)
					EncryptedStr2 = fmt.Sprintf("PLAIN_%s",mergedEncryptedTransaction)
					errorCount++;
				} else {
					EncryptedStr2 = SerilizeEncryption(dBytes)
					successCount++;
				}
			}
			
		encryptTransactions2[i] = EncryptedStr2
		encryptTransactions[i] = encryptedTrans
	}
		fmt.Printf("Performed Encrypted Transactions with successfull encryptions: %d, Errors: %d",successCount,errorCount)
		fmt.Println("All transactions are Proccessed")
		return encryptTransactions,encryptTransactions2,nil
}