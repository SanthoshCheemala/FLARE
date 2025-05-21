package storage

import (
	"database/sql"
	"fmt"
	"log"

	// "strings"
	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct{
	Data map[string]string
}

// OpenDatabase opens and returns a connection to the SQLite database at the given path
func OpenDatabase(DBpath string) *sql.DB {
	db, err := sql.Open("sqlite3", DBpath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func DisplayColumns(db *sql.DB, tableName string, columns []string, limit int) {
	// Build columns string for SQL query
	cols := ""
	for i, v := range columns {
		if i != len(columns) - 1 {
			cols += v + ", "
		} else {
			cols += v
		}
	}
	
	// Use fmt.Sprintf for proper string formatting and parameter binding
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d", cols, tableName, limit)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Get column names from the result set
	columnNames, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	
	// Create dynamic scan destinations
	values := make([]interface{}, len(columnNames))
	valuePtrs := make([]interface{}, len(columnNames))
	for i := range columnNames {
		valuePtrs[i] = &values[i]
	}
	
	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatal(err)
		}
		
		for i, col := range columnNames {
			val := values[i]
			// Handle different types of values
			switch v := val.(type) {
			case []byte:
				fmt.Printf("%s: %s ", col, string(v))
			default:
				fmt.Printf("%s: %v ", col, v)
			}
		}
		fmt.Println()
	}
	
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}

