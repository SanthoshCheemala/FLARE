package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	// "time"
	"github.com/SanthoshCheemala/FLARE.git/internal/crypto"
	"github.com/SanthoshCheemala/FLARE.git/internal/storage"
)

func main() {

	cols := flag.String("columns", "type,amount", "Comma-separated list of columns to encrypt")
	mergedCols := flag.String("columns-merge", "", "Comma-separated list of columns to merge for encryption")
	// encrypt := flag.Bool("encrypt", false, "Enable encryption mode")
	// decrypt := flag.Bool("decrypt", false, "Enable decryption mode")
	limit := flag.Int("LIMIT", 50, "Number of rows to process from the beginning")


	flag.Parse()

	if err := validateFlags(*cols, *mergedCols, *limit); err != nil {
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
	fmt.Println("- Row limit:", *limit)
	
	processData(columns, mergedColumns, *limit)
}

func validateFlags( cols, mergedCols string, limit int) error {

	if cols == "" {
		return fmt.Errorf("must specify at least one column with -columns")
	}

	if limit < 1 {
		return fmt.Errorf("LIMIT must be a positive integer")
	}

	return nil
}

func processData(columns, ColumnsTables []string, limit int) {
	db := storage.OpenDatabase("data/transactions.db")
	data := storage.RetriveData(db,"finanical_transactions",columns,ColumnsTables,limit)
	Intersection,err := crypto.Laconic_PSI(data[0:1],data[0:1],"data/tree.db")
	if err != nil {
		fmt.Print(err)
	}
	Intersection1,err := crypto.Laconic_PSI(data[0:1],data[1:2],"data/tree2.db")
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println("length: ",len(Intersection),len(Intersection))
	fmt.Println(Intersection,Intersection1)
	// storage.CreateDatabase(transactions,"LE_Table",columns,"data/encrypt.db")
}