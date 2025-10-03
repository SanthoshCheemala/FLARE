//go:build analytics
// +build analytics

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
	// Analytics command line flags
	cols := flag.String("columns", "type,amount", "Comma-separated list of columns to encrypt")
	mergedCols := flag.String("columns-merge", "", "Comma-separated list of columns to merge for encryption")
	limit := flag.Int("LIMIT", 50, "Number of rows to process from the beginning")
	outputDir := flag.String("output-dir", "data", "Directory for output files")
	enableAdvancedAnalytics := flag.Bool("advanced-analytics", true, "Enable advanced analytics and reporting")
	reportFormat := flag.String("report-format", "html", "Report format: html, json, or both")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	ringDimension := flag.Int("ring-dimension", 256, "Lattice ring dimension (256, 512, 1024, 2048)")

	flag.Parse()

	if err := validateFlags(*cols, *mergedCols, *limit, *outputDir, *ringDimension); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	columns := strings.Split(*cols, ",")
	var mergedColumns []string
	if *mergedCols != "" {
		mergedColumns = strings.Split(*mergedCols, ",")
	}

	fmt.Println("🔬 FLARE - Analytics Build")
	fmt.Println("================================================")
	fmt.Println("Configuration:")
	fmt.Printf("  📊 Columns to process: %v\n", columns)
	if len(mergedColumns) > 0 {
		fmt.Printf("  🔗 Merged columns: %v\n", mergedColumns)
	}
	fmt.Printf("  📈 Row limit: %d\n", *limit)
	fmt.Printf("  📁 Output directory: %s\n", *outputDir)
	fmt.Printf("  🔬 Advanced analytics: %t\n", *enableAdvancedAnalytics)
	fmt.Printf("  📋 Report format: %s\n", *reportFormat)
	fmt.Printf("  🔐 Ring dimension: %d\n", *ringDimension)
	fmt.Println()

	success := processDataAnalytics(columns, mergedColumns, *limit, *outputDir, *enableAdvancedAnalytics, *reportFormat, *verbose, *ringDimension)

	if success {
		fmt.Println("✅ FLARE analytics execution completed successfully!")
		fmt.Printf("📊 Check %s/ for generated reports\n", *outputDir)
	} else {
		fmt.Println("❌ FLARE analytics execution completed with errors")
		os.Exit(1)
	}
}

func validateFlags(cols, mergedCols string, limit int, outputDir string, ringDimension int) error {
	if cols == "" {
		return fmt.Errorf("must specify at least one column with -columns")
	}
	if limit < 1 {
		return fmt.Errorf("LIMIT must be a positive integer")
	}
	if outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	if ringDimension != 256 && ringDimension != 512 && ringDimension != 1024 && ringDimension != 2048 {
		return fmt.Errorf("ring dimension must be one of: 256, 512, 1024, 2048")
	}
	return nil
}

func processDataAnalytics(columns, columnsTables []string, limit int, outputDir string, enableAdvancedAnalytics bool, reportFormat string, verbose bool, ringDimension int) bool {
	dbPath := filepath.Join("data", "transactions.db")
	treeDbPath := filepath.Join(outputDir, "tree.db")

	if verbose {
		fmt.Printf("🗄️  Opening database: %s\n", dbPath)
	}

	db := storage.OpenDatabase(dbPath)
	if db == nil {
		fmt.Printf("❌ Failed to open database: %s\n", dbPath)
		return false
	}
	defer db.Close()

	if verbose {
		fmt.Println("📥 Retrieving transaction data...")
	}

	data := storage.RetriveData(db, "finanical_transactions", columns, columnsTables, limit)
	if len(data) == 0 {
		fmt.Println("❌ No data retrieved from database")
		return false
	}

	fmt.Printf("✅ Retrieved %d transaction records\n", len(data))

	// Split data into client and server sets
	clientSize := len(data) / 3
	if clientSize == 0 {
		clientSize = 1
	}

	clientData := data[0:clientSize]
	serverData := data[0:limit]

	fmt.Printf("🔄 Processing PSI with %d client records and %d server records\n", len(clientData), len(serverData))

	if verbose {
		fmt.Printf("🌳 Using tree database: %s\n", treeDbPath)
	}

	intersection, err := crypto.Laconic_PSI_WithAnalyticsCustom(clientData, serverData, treeDbPath, ringDimension)
	if err != nil {
		fmt.Printf("❌ PSI execution failed: %v\n", err)
		return false
	}

	fmt.Printf("🎯 PSI completed successfully!\n")
	fmt.Printf("📊 Found %d intersections\n", len(intersection))

	// Print intersection results if verbose
	if verbose && len(intersection) > 0 {
		fmt.Println("\n🔍 Intersection Results:")
		for i, trans := range intersection {
			fmt.Printf("  %d: %v\n", i+1, trans.Data)
			if i >= 4 {
				fmt.Printf("  ... and %d more\n", len(intersection)-5)
				break
			}
		}
	}

	// Display report locations with correct paths
	fmt.Println("\n📋 Generated Reports:")
	
	if reportFormat == "html" || reportFormat == "both" {
		htmlPath := filepath.Join(outputDir, "flare_psi_advanced_report.html")
		fmt.Printf("  🌐 Interactive Dashboard: %s\n", htmlPath)
	}
	
	if reportFormat == "json" || reportFormat == "both" {
		jsonPath := filepath.Join(outputDir, "flare_psi_statistics.json")
		fmt.Printf("  📊 Statistics JSON: %s\n", jsonPath)
	}
	
	if enableAdvancedAnalytics {
		fmt.Printf("  ⚡ Performance Profile: %s/performance_profile.json\n", outputDir)
		fmt.Printf("  🛡️  Security Assessment: %s/security_assessment.json\n", outputDir)
		fmt.Printf("  🔧 Optimization Report: %s/optimization_recommendations.json\n", outputDir)
	}
	
	return true
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `FLARE - Analytics Build with Advanced Features

Usage: %s [OPTIONS]

Options:
`, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Advanced analytics with custom output and ring dimension
  %s -columns="user_id,transaction_type" -output-dir="results" -advanced-analytics=true -ring-dimension=512

  # Verbose execution with JSON reports and higher security
  %s -columns="type,amount,timestamp" -LIMIT=200 -verbose=true -report-format="json" -ring-dimension=1024

  # Production mode with merged columns and default security
  %s -columns="type,amount" -columns-merge="user_id,timestamp" -LIMIT=1000 -report-format="both" -ring-dimension=256
`, os.Args[0], os.Args[0], os.Args[0])
	}
}