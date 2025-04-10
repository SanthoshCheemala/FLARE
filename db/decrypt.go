package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/LE"
	"github.com/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/matrix"
)

// DecryptDatabase reads data from an encrypted database and returns decrypted transactions
func DecryptDatabase(dbPath string, tableName string, columns []string, limit int, 
	leParams *LE.LE, secretKey *matrix.Vector, treeDB *sql.DB) ([]Transaction, error) {
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	
	// Build query to select specified columns with limit
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d", 
		strings.Join(columns, ", "), tableName, limit)
	
	fmt.Println("Executing decryption query:", query)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()
	
	var transactions []Transaction
	
	rowIdx := 0
	for rows.Next() {
		// Create scan destinations
		scanDest := make([]interface{}, len(columns))
		for i := range scanDest {
			var val string
			scanDest[i] = &val
		}
		
		// Scan the row into destinations
		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		
		// Create a new transaction
		trans := Transaction{
			Data: make(map[string]string),
		}
		
		// Decrypt each field
		for i, col := range columns {
			encryptedStr := *scanDest[i].(*string)
			
			// Calculate the ID used for encryption
			fieldID := uint64(rowIdx*len(columns) + i)
			
			// Skip decryption for empty fields
			if encryptedStr == "" {
				trans.Data[col] = ""
				continue
			}
			
			// Decrypt the data
			decryptedStr, err := LE.Decrypt(leParams, encryptedStr, secretKey, treeDB, fieldID)
			if err != nil {
				return nil, fmt.Errorf("error decrypting field %s: %w", col, err)
			}
			
			// Store the decrypted data
			trans.Data[col] = decryptedStr
		}
		
		transactions = append(transactions, trans)
		rowIdx++
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return transactions, nil
}
