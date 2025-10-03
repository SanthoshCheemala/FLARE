//go:build !analytics && !debug
// +build !analytics,!debug

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SanthoshCheemala/FLARE/internal/crypto"
	"github.com/SanthoshCheemala/FLARE/internal/storage"
)

func main() {
	cols := flag.String("columns", "type,amount", "Comma-separated list of columns to encrypt")
	limit := flag.Int("LIMIT", 2, "Number of rows to process from the beginning")

	flag.Parse()

	if err := validateFlags(*cols, *limit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	columns := strings.Split(*cols, ",")
	
	// --- IGNORE --- 
	fmt.Println("🚀 FLARE - Lattice-based Private Set Intersection")
	fmt.Println("Configuration:")
	fmt.Printf("  📊 Columns: %v\n", columns)
	fmt.Printf("  📈 Limit: %d\n", *limit)
	fmt.Println()

	success := processData(columns, *limit)

	if success {
		fmt.Println("FLARE execution completed successfully!")
	} else {
		fmt.Println("FLARE execution failed")
		os.Exit(1)
	}
}

func validateFlags(cols string, limit int) error {
	if cols == "" {
		return fmt.Errorf("must specify at least one column with -columns")
	}
	if limit < 1 {
		return fmt.Errorf("LIMIT must be a positive integer")
	}
	return nil
}

func processData(columns []string, limit int) bool {
	dbPath := filepath.Join("data", "transactions.db")

	db := storage.OpenDatabase(dbPath)
	if db == nil {
		fmt.Printf("❌ Failed to open database: %s\n", dbPath)
		return false
	}
	defer db.Close()

	data := storage.RetriveData(db, "finanical_transactions", columns, nil, limit)
	if len(data) == 0 {
		fmt.Println("No data retrieved from database")
		return false
	}

	fmt.Printf("Retrieved %d records\n", len(data))

	clientSize := len(data) / 2
	if clientSize == 0 {
		clientSize = 1
	}

	clientData := data[0:clientSize]
	serverData := data[0:limit]

	intersection, err := crypto.Laconic_PSI(clientData, serverData, "data/tree.db")
	
	if err != nil {
		fmt.Printf("❌ PSI failed: %v\n", err)
		return false
	}

	fmt.Printf("Found %d intersections\n", len(intersection))
	fmt.Printf("%v\n", intersection)
	return true
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `FLARE - Production PSI System

Usage: %s [OPTIONS]

Options:
`, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Example:
  %s -columns="type,amount" -LIMIT=100
`, os.Args[0])
	}
}