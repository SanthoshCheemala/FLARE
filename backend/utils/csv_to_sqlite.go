package utils

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// CSVToSQLiteConfig holds configuration for CSV to SQLite conversion
type CSVToSQLiteConfig struct {
	CSVFiles     []string // List of CSV file paths to convert
	OutputDBPath string   // Path where the SQLite database will be saved
	MaxWorkers   int      // Maximum number of concurrent workers (0 = use all CPU cores)
	BatchSize    int      // Number of rows to insert in a single transaction
	CreateTables bool     // Whether to create tables automatically
	DropExisting bool     // Whether to drop existing tables before creating new ones
}

// CSVConversionResult holds the result of a CSV conversion
type CSVConversionResult struct {
	FileName  string
	TableName string
	RowCount  int
	Error     error
}

// CSVToSQLiteConverter handles the conversion of CSV files to SQLite
type CSVToSQLiteConverter struct {
	config *CSVToSQLiteConfig
	db     *sql.DB
	mu     sync.Mutex
}

// NewCSVToSQLiteConverter creates a new converter instance
func NewCSVToSQLiteConverter(config *CSVToSQLiteConfig) (*CSVToSQLiteConverter, error) {
	// Set default values
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = runtime.NumCPU()
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 1000
	}
	if config.OutputDBPath == "" {
		config.OutputDBPath = "data/output.db"
	}

	// Ensure the data directory exists
	dbDir := filepath.Dir(config.OutputDBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", config.OutputDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Configure SQLite for better performance
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=10000",
		"PRAGMA temp_store=MEMORY",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %v", err)
		}
	}

	return &CSVToSQLiteConverter{
		config: config,
		db:     db,
	}, nil
}

// Close closes the database connection
func (c *CSVToSQLiteConverter) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Convert processes all CSV files and converts them to SQLite tables
func (c *CSVToSQLiteConverter) Convert() ([]CSVConversionResult, error) {
	results := make([]CSVConversionResult, len(c.config.CSVFiles))

	// Create a semaphore to limit concurrent workers
	semaphore := make(chan struct{}, c.config.MaxWorkers)
	var wg sync.WaitGroup

	fmt.Printf("Starting conversion with %d workers (CPU cores available: %d)\n",
		c.config.MaxWorkers, runtime.NumCPU())

	// Process each CSV file concurrently
	for i, csvFile := range c.config.CSVFiles {
		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := c.convertFile(filePath)
			results[index] = result

			if result.Error != nil {
				fmt.Printf("❌ Failed to convert %s: %v\n", result.FileName, result.Error)
			} else {
				fmt.Printf("✅ Converted %s -> %s (%d rows)\n",
					result.FileName, result.TableName, result.RowCount)
			}
		}(i, csvFile)
	}

	// Wait for all conversions to complete
	wg.Wait()

	fmt.Println("\n🎉 Conversion completed!")
	return results, nil
}

// convertFile converts a single CSV file to a SQLite table
func (c *CSVToSQLiteConverter) convertFile(csvPath string) CSVConversionResult {
	result := CSVConversionResult{
		FileName: filepath.Base(csvPath),
	}

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open CSV file: %v", err)
		return result
	}
	defer file.Close()

	// Read CSV
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		result.Error = fmt.Errorf("failed to read CSV headers: %v", err)
		return result
	}

	// Generate table name from filename
	tableName := sanitizeTableName(strings.TrimSuffix(result.FileName, filepath.Ext(result.FileName)))
	result.TableName = tableName

	// Create table
	if err := c.createTable(tableName, headers); err != nil {
		result.Error = err
		return result
	}

	// Insert data in batches
	rowCount := 0
	batch := [][]string{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Error = fmt.Errorf("failed to read CSV record: %v", err)
			return result
		}

		batch = append(batch, record)

		// Insert batch when it reaches the configured size
		if len(batch) >= c.config.BatchSize {
			if err := c.insertBatch(tableName, headers, batch); err != nil {
				result.Error = err
				return result
			}
			rowCount += len(batch)
			batch = [][]string{}
		}
	}

	// Insert remaining records
	if len(batch) > 0 {
		if err := c.insertBatch(tableName, headers, batch); err != nil {
			result.Error = err
			return result
		}
		rowCount += len(batch)
	}

	result.RowCount = rowCount
	return result
}

// createTable creates a table in the database
func (c *CSVToSQLiteConverter) createTable(tableName string, columns []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.config.CreateTables {
		return nil
	}

	// Drop table if requested
	if c.config.DropExisting {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		if _, err := c.db.Exec(dropSQL); err != nil {
			return fmt.Errorf("failed to drop table: %v", err)
		}
	}

	// Create table with all columns as TEXT type
	columnDefs := make([]string, len(columns))
	for i, col := range columns {
		columnDefs[i] = fmt.Sprintf("%s TEXT", sanitizeColumnName(col))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
		tableName, strings.Join(columnDefs, ", "))

	if _, err := c.db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// insertBatch inserts a batch of rows into the database
func (c *CSVToSQLiteConverter) insertBatch(tableName string, headers []string, batch [][]string) error {
	if len(batch) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Prepare insert statement
	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sanitizedHeaders := make([]string, len(headers))
	for i, h := range headers {
		sanitizedHeaders[i] = sanitizeColumnName(h)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(sanitizedHeaders, ", "),
		strings.Join(placeholders, ", "))

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %v", err)
	}
	defer stmt.Close()

	// Insert all rows in the batch
	for _, record := range batch {
		// Convert []string to []interface{}
		values := make([]interface{}, len(record))
		for i, v := range record {
			values[i] = v
		}

		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("failed to insert record: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// sanitizeTableName cleans table name for SQLite
func sanitizeTableName(name string) string {
	// Replace invalid characters with underscores
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)

	// Ensure it doesn't start with a number
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "tbl_" + name
	}

	return name
}

// sanitizeColumnName cleans column name for SQLite
func sanitizeColumnName(name string) string {
	// Replace invalid characters with underscores
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)

	// Ensure it doesn't start with a number
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "col_" + name
	}

	return name
}

// ConvertCSVToSQLite is a convenience function for quick conversions
func ConvertCSVToSQLite(csvFiles []string, outputDBPath string) error {
	config := &CSVToSQLiteConfig{
		CSVFiles:     csvFiles,
		OutputDBPath: outputDBPath,
		MaxWorkers:   0, // Use all CPU cores
		BatchSize:    1000,
		CreateTables: true,
		DropExisting: true,
	}

	converter, err := NewCSVToSQLiteConverter(config)
	if err != nil {
		return err
	}
	defer converter.Close()

	results, err := converter.Convert()
	if err != nil {
		return err
	}

	// Check for any errors in results
	for _, result := range results {
		if result.Error != nil {
			return fmt.Errorf("conversion failed for %s: %v", result.FileName, result.Error)
		}
	}

	return nil
}
