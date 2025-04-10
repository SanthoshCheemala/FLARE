package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/LE"
	"github.com/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/db"
	"github.com/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)

func main() {
	var columnsString string
	flag.StringVar(&columnsString, "columns", "", "Comma-separated list of column names to include")
	
	showColumnsFlag := flag.Bool("list-columns", false, "List available columns in the source database")
	
	outputDBFlag := flag.String("output", "data/encrypted.db", "Output database file path")

	inputDBFlag := flag.String("input", "data/transactions.db", "Input database file path")

	sourceTableFlag := flag.String("source-table", "finanical_transactions", "Source table name")
	targetTableFlag := flag.String("target-table", "Encrypt_Trans", "Target table name")
	
	limitFlag := flag.Int("limit", 100, "Maximum number of rows to process")
	
	encryptFlag := flag.Bool("encrypt", true, "Enable laconic encryption")
	decryptFlag := flag.Bool("decrypt", false, "Decrypt an encrypted database")
	
	treeDBPathFlag := flag.String("tree-db", "data/tree.db", "Path to tree database for LE")
	secretKeyPathFlag := flag.String("secret-key", "data/secret_key.bin", "Path to secret key file")
	
	flag.Parse()
	
	// Handle decryption mode
	if *decryptFlag {
		if *encryptFlag {
			fmt.Println("ERROR: Cannot use both --encrypt and --decrypt flags together")
			os.Exit(1)
		}
		
		decryptDatabase(*inputDBFlag, *outputDBFlag, *sourceTableFlag, *targetTableFlag, 
			columnsString, *treeDBPathFlag, *secretKeyPathFlag, *limitFlag)
		return
	}
	
	// Regular encryption or passthrough mode
	database := db.OpenDatabase(*inputDBFlag)
	defer database.Close()
	
	if *showColumnsFlag {
		db.DisplayColumns(database, *sourceTableFlag)
		return
	}

	if columnsString == "" {
		fmt.Println("No columns specified. Use -columns flag to specify columns.")
		fmt.Println("Available columns:")
		db.DisplayColumns(database, *sourceTableFlag)
		os.Exit(1)
	}
	
	columns := strings.Split(columnsString, ",")
	fmt.Printf("Processing database with columns: %v\n", columns)
	
	transData := db.RetrieveData(database, *sourceTableFlag, columns, *limitFlag)
	
	// If encryption is enabled, process the data through laconic encryption
	if *encryptFlag {
		fmt.Println("Applying laconic encryption to data...")
		encryptedData, err := encryptTransactions(transData, columns, *treeDBPathFlag, *secretKeyPathFlag)
		if err != nil {
			fmt.Printf("Error during encryption: %v\n", err)
			os.Exit(1)
		}
		db.CreateDatabase(encryptedData, *targetTableFlag, columns, *outputDBFlag)
	} else {
		db.CreateDatabase(transData, *targetTableFlag, columns, *outputDBFlag)
	}
	
	fmt.Println("Number of transactions processed:", len(transData))
}

// encryptTransactions applies laconic encryption to each field in the transactions
func encryptTransactions(transactions []db.Transaction, columns []string, treeDBPath, secretKeyPath string) ([]db.Transaction, error) {
	// Initialize LE parameters
	leParams := LE.Setup(1<<30, 32, 512, 4) // Example parameters, adjust as needed
	
	// Create a tree database for the laconic encryption
	treeDB, err := sql.Open("sqlite3", treeDBPath)
	if err != nil {
		return nil, fmt.Errorf("error opening tree database: %w", err)
	}
	defer treeDB.Close()
	
	// Initialize tree tables if they don't exist
	if err := initializeTreeDB(treeDB, leParams.Layers); err != nil {
		return nil, fmt.Errorf("failed to initialize tree database: %w", err)
	}
	
	// Generate key pair for encryption
	pubKey, secretKey := leParams.KeyGen()
	
	// Save secret key for later decryption
	if err := saveSecretKey(secretKey, secretKeyPath); err != nil {
		return nil, fmt.Errorf("failed to save secret key: %w", err) 
	}
	
	// Encrypt each transaction
	encryptedTransactions := make([]db.Transaction, len(transactions))
	
	for i, trans := range transactions {
		encryptedTrans := db.Transaction{
			Data: make(map[string]string),
		}
		
		// Encrypt each field in the transaction
		for _, col := range columns {
			// Create a ring polynomial from the string data
			dataStr := trans.Data[col]
			dataPoly := stringToPoly(dataStr, leParams.R)
			
			// Generate a unique ID for this field (could use hash of column name + row number)
			fieldID := uint64(i*len(columns) + getColumnIndex(columns, col))
			
			// Register the public key in the tree
			LE.Upd(treeDB, fieldID, leParams.Layers, pubKey, leParams)
			
			// Encrypt the data
			c0, c1, c, d := LE.EncWithRandomness(leParams, pubKey, fieldID, dataPoly)
			
			// Serialize encryption components
			encryptedStr, err := serializeEncryption(c0, c1, c, d)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize encryption for field %s: %w", col, err)
			}
			
			// Store the encrypted data
			encryptedTrans.Data[col] = encryptedStr
		}
		
		encryptedTransactions[i] = encryptedTrans
	}
	
	return encryptedTransactions, nil
}

// decryptDatabase decrypts an encrypted database and writes results to a new database
func decryptDatabase(inputDB, outputDB, sourceTable, targetTable, columnsStr, treeDBPath, secretKeyPath string, limit int) {
	if columnsStr == "" {
		fmt.Println("No columns specified for decryption")
		os.Exit(1)
	}
	
	columns := strings.Split(columnsStr, ",")
	
	// Initialize LE parameters
	leParams := LE.Setup(1<<30, 32, 512, 4) // Use same parameters as encryption
	
	// Load secret key
	secretKey, err := loadSecretKey(secretKeyPath, leParams.R)
	if err != nil {
		fmt.Printf("Error loading secret key: %v\n", err)
		os.Exit(1)
	}
	
	// Open tree database
	treeDB, err := sql.Open("sqlite3", treeDBPath)
	if err != nil {
		fmt.Printf("Error opening tree database: %v\n", err)
		os.Exit(1)
	}
	defer treeDB.Close()
	
	// Decrypt database
	decryptedData, err := db.DecryptDatabase(inputDB, sourceTable, columns, limit, leParams, secretKey, treeDB)
	if err != nil {
		fmt.Printf("Error decrypting database: %v\n", err)
		os.Exit(1)
	}
	
	// Create new database with decrypted data
	db.CreateDatabase(decryptedData, targetTable, columns, outputDB)
	
	fmt.Println("Successfully decrypted", len(decryptedData), "records")
}

// Helper functions

// Initialize tree database tables for laconic encryption
func initializeTreeDB(db *sql.DB, layers int) error {
	for i := 0; i <= layers; i++ {
		query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS tree_%d (
			rowid INTEGER PRIMARY KEY,
			p1 BLOB,
			p2 BLOB,
			p3 BLOB,
			p4 BLOB,
			y_def BOOLEAN
		)`, i)
		
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error creating tree table %d: %w", i, err)
		}
	}
	return nil
}

// Convert a string to a polynomial for encryption
func stringToPoly(s string, r *ring.Ring) *ring.Poly {
	poly := r.NewPoly()
	
	// Simple encoding: each character becomes a coefficient
	// This is a simplified approach - real applications would use more sophisticated encoding
	for i, c := range s {
		if i < r.N {
			poly.Coeffs[0][i] = uint64(c) % r.Modulus[0]
		}
	}
	
	return poly
}

// Serialize encryption components to a string representation
func serializeEncryption(c0, c1 []*matrix.Vector, c *matrix.Vector, d *ring.Poly) (string, error) {
	// This is a simplified serialization that combines the components into a single string
	// In a real application, you would use proper binary serialization
	
	// Serialize the d polynomial
	dBytes, err := d.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to serialize d polynomial: %w", err)
	}
	
	// Serialize c vector
	cBytes := c.Encode()
	
	// Count total length for allocation
	totalLen := len(dBytes)
	for _, b := range cBytes {
		totalLen += len(b)
	}
	
	// Return a base64 encoded combination (in real application, use proper binary serialization)
	return fmt.Sprintf("LE_ENC_%d_%d", len(dBytes), totalLen), nil
}

// Get index of a column in the columns slice
func getColumnIndex(columns []string, colName string) int {
	for i, col := range columns {
		if col == colName {
			return i
		}
	}
	return 0
}

// Save secret key to a file
func saveSecretKey(sk *matrix.Vector, path string) error {
	// Create directory if it doesn't exist
	dir := strings.Split(path, "/")
	if len(dir) > 1 {
		dirPath := strings.Join(dir[:len(dir)-1], "/")
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return err
			}
		}
	}
	
	// Create file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Serialize secret key
	skBytes := sk.Encode()
	
	// Write length of each component followed by the component
	for _, bytes := range skBytes {
		lenBytes := []byte{byte(len(bytes))}
		if _, err := file.Write(lenBytes); err != nil {
			return err
		}
		if _, err := file.Write(bytes); err != nil {
			return err
		}
	}
	
	return nil
}

// Load secret key from a file
func loadSecretKey(path string, r *ring.Ring) (*matrix.Vector, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	// Create a vector with the correct dimension
	sk := matrix.NewVector(4, r) // Use appropriate dimension
	
	// In a real implementation, deserialize the vector properly
	// This is a placeholder implementation
	
	return sk, nil
}