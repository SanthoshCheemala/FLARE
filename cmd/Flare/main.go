package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/SanthoshCheemala/FLARE.git/internal/crypto"
	"github.com/SanthoshCheemala/FLARE.git/internal/storage"
)

func main() {

	cols := flag.String("columns", "type,amount", "Comma-separated list of columns to encrypt")
	mergedCols := flag.String("columns-merge", "", "Comma-separated list of columns to merge for encryption")
	encrypt := flag.Bool("encrypt", false, "Enable encryption mode")
	decrypt := flag.Bool("decrypt", false, "Enable decryption mode")
	limit := flag.Int("LIMIT", 100, "Number of rows to process from the beginning")


	flag.Parse()

	if err := validateFlags(*encrypt, *decrypt, *cols, *mergedCols, *limit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	columns := strings.Split(*cols, ",")
	var mergedColumns []string
	if *mergedCols != "" {
		mergedColumns = strings.Split(*mergedCols, ",")
	}

	fmt.Println("Configuration:")
	fmt.Println("- Columns to process:", columns)
	if len(mergedColumns) > 0 {
		fmt.Println("- Merged columns:", mergedColumns)
	}
	fmt.Printf("- Mode: %s\n", getMode(*encrypt, *decrypt))
	fmt.Println("- Row limit:", *limit)
	
	processData(columns, mergedColumns, *encrypt, *decrypt, *limit)
}

func validateFlags(encrypt, decrypt bool, cols, mergedCols string, limit int) error {
	if encrypt && decrypt {
		return fmt.Errorf("cannot use both -encrypt and -decrypt flags simultaneously")
	}

	if !encrypt && !decrypt {
		return fmt.Errorf("must specify either -encrypt or -decrypt")
	}

	if cols == "" {
		return fmt.Errorf("must specify at least one column with -columns")
	}

	if limit < 1 {
		return fmt.Errorf("LIMIT must be a positive integer")
	}

	return nil
}

func getMode(encrypt, decrypt bool) string {
	if encrypt {
		return "Encryption"
	}
	return "Decryption"
}

func processData(columns, mergedColumns []string, encrypt, decrypt bool, limit int) {
	db := storage.OpenDatabase("data/transactions.db")
	data := storage.RetriveData(db,"finanical_transactions",columns,mergedColumns,limit)
	var UnEncryptedData []byte
	start := time.Now()
	ecnryptedTransactions,err  := crypto.EncryptTransactions(data,columns,"data/tree.db","data/secret.bin")
	if err != nil{
		log.Fatal(err)
	}
	end := time.Now()
	totalTime := end.Sub(start)
	var EncryptedData []byte
	fmt.Println("Total Time to Encrypt the data is ",totalTime)
	for _,v := range data{
		fmt.Println(v)
		for _,c := range v.Data["amount"]{
			UnEncryptedData = append(UnEncryptedData,byte(c))
		}
	}
	for _,v := range ecnryptedTransactions{
		fmt.Println(v.Data["amount"])
		for _,c := range v.Data["amount"]{
			EncryptedData = append(EncryptedData, byte(c))
		}
	}
	fmt.Println(len(UnEncryptedData),"/",len(EncryptedData))
	
}