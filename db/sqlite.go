package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Transaction represents a database row with dynamic columns
type Transaction struct {
	Data map[string]string
}

// RetrieveData fetches data from a SQLite database with specified columns
func RetrieveData(db *sql.DB, tableName string, columns []string, limit int) []Transaction {
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d", 
		strings.Join(columns, ", "), tableName, limit)
	
	fmt.Println("Executing query:", query)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var data []Transaction
	for rows.Next() {
		scanValues := make([]interface{}, len(columns))
		scanPointers := make([]interface{}, len(columns))
		for i := range columns {
			scanPointers[i] = &scanValues[i]
		}
		
		err = rows.Scan(scanPointers...)
		if err != nil {
			log.Fatal(err)
		}
		
		rowData := make(map[string]string)
		for i, col := range columns {
			val := scanPointers[i].(*interface{})
			switch v := (*val).(type) {
			case []byte:
				rowData[col] = string(v)
			case string:
				rowData[col] = v
			case nil:
				rowData[col] = ""
			default:
				rowData[col] = fmt.Sprintf("%v", v)
			}
		}
		
		data = append(data, Transaction{Data: rowData})
		
		if err = rows.Err(); err != nil {
			log.Fatal(err)
		}
	}
	return data
}

// CreateDatabase creates a new SQLite database with specified columns and data
func CreateDatabase(trans []Transaction, tableName string, columns []string, outputFile string) {
	db, err := sql.Open("sqlite3", outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Create table with dynamic columns
	columnDefs := make([]string, len(columns))
	for i, col := range columns {
		columnDefs[i] = fmt.Sprintf("%s TEXT", col)
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);",
		tableName, strings.Join(columnDefs, ", "))
	fmt.Println("Creating table with SQL:", createTableSQL)

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Insert data
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))
	fmt.Println("Inserting with SQL:", insertSQL)

	for i := 0; i < len(trans); i++ {
		// Extract values from the transaction map in the same order as columns
		values := make([]interface{}, len(columns))
		for j, col := range columns {
			values[j] = trans[i].Data[col]
		}

		_, err = db.Exec(insertSQL, values...)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetAvailableColumns retrieves the column information from a table
func GetAvailableColumns(db *sql.DB, tableName string) ([]string, []string) {
	// Query to get column information from SQLite
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var columnNames []string
	var columnTypes []string
	
	var id int
	var name string
	var dataType string
	var notNull int
	var defaultVal interface{}
	var pk int
	
	for rows.Next() {
		if err := rows.Scan(&id, &name, &dataType, &notNull, &defaultVal, &pk); err != nil {
			log.Fatal(err)
		}
		columnNames = append(columnNames, name)
		columnTypes = append(columnTypes, dataType)
	}
	
	return columnNames, columnTypes
}

// OpenDatabase opens a SQLite database connection
func OpenDatabase(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// DisplayColumns prints available columns in a database table
func DisplayColumns(db *sql.DB, tableName string) {
	names, types := GetAvailableColumns(db, tableName)
	
	fmt.Printf("Available columns in %s:\n", tableName)
	fmt.Println("-------------------------------------------")
	
	for i := 0; i < len(names); i++ {
		fmt.Printf("- %s (%s)\n", names[i], types[i])
	}
	
	fmt.Println("\nUsage example:")
	fmt.Printf("  go run main.go -columns=%s -output=mydb.db -limit=500\n\n", 
		strings.Join(names[:min(2, len(names))], ","))
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
