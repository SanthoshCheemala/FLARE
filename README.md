# FLARE - Lattice Encryption Framework

## Overview

Flare is a database management toolkit focused on SQLite operations, providing tools for:

1. Converting CSV files to SQL statements (csv2sql)
2. Database encryption and processing with Lattice Encryption (LE)
3. Transaction data extraction and transformation

## Features

### CSV to SQL Conversion (csv2sql)
- Converts CSV files to SQLite-compatible SQL statements
- Validates CSV integrity during conversion
- Automatically cleans column headers for SQL compatibility
- Creates SQL files ready to be imported with SQLite's `.read` command

### Database Operations (main)
- Extract specific columns from SQLite databases
- Create new databases with selected data
- Support for encrypted databases
- Configuration through command-line arguments

### Lattice Encryption (LE)
- Matrix-based encryption for database content
- Tree-based storage of encrypted values
- Support for complex encryption parameters
- Full encryption and decryption of database records

## Project Structure

The project has been reorganized into a modular structure:

```
/Users/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare/
├── cmd/
│   └── flare/          # Command-line application
│       └── main.go     # Application entry point
├── internal/
│   ├── crypto/         # Encryption functionality
│   ├── storage/        # Database operations
│   └── utils/          # Utility functions
└── pkg/
    └── le/             # Reusable Lattice Encryption package
```

## Usage

### Basic Database Operations
```
# List columns in a database table
go run main.go -input=data/transactions.db -source-table=financial_transactions -list-columns

# Extract specific columns without encryption
go run main.go -input=data/transactions.db -output=data/extracted.db -columns=id,amount,date -encrypt=false

# Apply encryption to specific columns
go run main.go -input=data/transactions.db -output=data/encrypted.db -columns=id,amount,date -encrypt
```

### Encryption & Decryption
```
# Encrypt a database
go run main.go -input=data/transactions.db -output=data/encrypted.db -columns=id,amount,date -encrypt -tree-db=data/tree.db -secret-key=data/secret.key

# Decrypt a database
go run main.go -input=data/encrypted.db -output=data/decrypted.db -columns=id,amount,date -decrypt -tree-db=data/tree.db -secret-key=data/secret.key
```

### CSV to SQL Conversion
```
# Convert CSV to SQL
go run csv2sql.go -f data/transactions.csv -t financial_transactions
```

## Running the Application

To run the application:

```bash
cd /Users/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/Flare
go run cmd/flare/main.go -columns=type,amount,nameDest -encrypt -output=data/encrypted.db
```

Use the `-list-columns` flag to see available columns in the source database.