package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/SanthoshCheemala/FLARE.git/internal/crypto"
	"github.com/SanthoshCheemala/FLARE.git/internal/storage"
	"github.com/SanthoshCheemala/FLARE.git/internal/utils"
)

/* Main function that handles command line arguments and controls the data flow */
func main() {
	var columnsString string
	flag.StringVar(&columnsString, "columns", "", "Comma-separated list of column names to include")
	
	defaultColumnsMsg := "newbalanceDest,isFraud,isFlaggedFraud,type,amount,nameDest"
	
	showColumnsFlag := flag.Bool("list-columns", false, "List available columns in the source database")
	outputDBFlag := flag.String("output", "data/encrypted.db", "Output database file path")
	inputDBFlag := flag.String("input", "data/transactions.db", "Input database file path")
	sourceTableFlag := flag.String("source-table", "finanical_transactions", "Source table name")
	targetTableFlag := flag.String("target-table", "Encrypt_Trans", "Target table name")
	limitFlag := flag.Int("limit", 100, "Maximum number of rows to process")
	encryptFlag := flag.Bool("encrypt", false, "Enable laconic encryption")
	treeDBPathFlag := flag.String("tree-db", "data/tree.db", "Path to tree database for LE")
	secretKeyPathFlag := flag.String("secret-key", "data/secret_key.bin", "Path to secret key file")
	
	flag.Parse()
	
	database := storage.OpenDatabase(*inputDBFlag)
	defer database.Close()
	
	if *showColumnsFlag {
		storage.DisplayColumns(database, *sourceTableFlag)
		return
	}

	if columnsString == "" {
		fmt.Println("No columns specified. Use -columns flag to specify columns.")
		fmt.Println("Available columns:")
		storage.DisplayColumns(database, *sourceTableFlag)
		fmt.Printf("\nExample usage: go run main.go -columns=%s -output=mydb.db -limit=500\n", defaultColumnsMsg)
		os.Exit(1)
	}
	
	columns := strings.Split(columnsString, ",")
	fmt.Printf("Processing database with columns: %v\n", columns)
	
	transData := storage.RetrieveData(database, *sourceTableFlag, columns, *limitFlag)
	
	if *encryptFlag {
		fmt.Println("Applying laconic encryption to data...")
		utils.EnsureDataDirectory()
		
		encryptedData, err := crypto.EncryptTransactions(transData, columns, *treeDBPathFlag, *secretKeyPathFlag)
		if err != nil {
			fmt.Printf("Error during encryption: %v\n", err)
			os.Exit(1)
		}
		storage.CreateDatabase(encryptedData, *targetTableFlag, columns, *outputDBFlag)
	} else {
		storage.CreateDatabase(transData, *targetTableFlag, columns, *outputDBFlag)
	}
	
	fmt.Println("Number of transactions processed:", len(transData))
}
