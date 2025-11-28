# CSV to SQLite Converter

A high-performance utility to convert CSV files to SQLite database using Go routines for parallel processing.

## Features

- 🚀 **Parallel Processing**: Utilizes all available CPU cores for maximum performance
- 📦 **Batch Processing**: Inserts data in configurable batches for optimal speed
- 🔄 **Concurrent Workers**: Process multiple CSV files simultaneously
- 💪 **Robust**: Handles large CSV files efficiently with transaction-based inserts
- ⚙️ **Configurable**: Customize workers, batch size, and more
- 🎯 **SQLite Optimized**: Uses WAL mode and other optimizations for fast writes

## Installation

The utility is already part of the FLARE project. Make sure you have the required dependencies:

```bash
go mod download
```

## Usage

### Method 1: Using the Command-Line Tool

```bash
# Convert specific CSV files
go run cmd/csv_converter/main.go -csv file1.csv,file2.csv -output data/mydb.db

# Convert all CSV files in a directory
go run cmd/csv_converter/main.go -dir /path/to/csv/files -output data/mydb.db

# Customize workers and batch size
go run cmd/csv_converter/main.go -dir ./csvfiles -workers 8 -batch 5000 -output data/mydb.db
```

#### Command-Line Options

- `-csv`: Comma-separated list of CSV file paths
- `-dir`: Directory containing CSV files (will process all .csv files)
- `-output`: Output SQLite database path (default: `data/output.db`)
- `-workers`: Number of concurrent workers (default: 0 = use all CPU cores)
- `-batch`: Number of rows per batch insert (default: 1000)
- `-drop`: Drop existing tables before creating new ones (default: true)

### Method 2: Using as a Library

```go
package main

import (
    "log"
    "github.com/SanthoshCheemala/FLARE/utils"
)

func main() {
    // Simple usage
    csvFiles := []string{
        "data/sales.csv",
        "data/customers.csv",
        "data/products.csv",
    }
    
    err := utils.ConvertCSVToSQLite(csvFiles, "data/mydata.db")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Method 3: Advanced Configuration

```go
package main

import (
    "log"
    "runtime"
    "github.com/SanthoshCheemala/FLARE/utils"
)

func main() {
    // Advanced configuration
    config := &utils.CSVToSQLiteConfig{
        CSVFiles: []string{
            "data/large_file1.csv",
            "data/large_file2.csv",
        },
        OutputDBPath: "data/output.db",
        MaxWorkers:   runtime.NumCPU() * 2, // Use more workers
        BatchSize:    5000,                  // Larger batch size
        CreateTables: true,
        DropExisting: true,
    }
    
    converter, err := utils.NewCSVToSQLiteConverter(config)
    if err != nil {
        log.Fatal(err)
    }
    defer converter.Close()
    
    results, err := converter.Convert()
    if err != nil {
        log.Fatal(err)
    }
    
    // Process results
    for _, result := range results {
        if result.Error != nil {
            log.Printf("Failed: %s - %v", result.FileName, result.Error)
        } else {
            log.Printf("Success: %s -> %s (%d rows)", 
                result.FileName, result.TableName, result.RowCount)
        }
    }
}
```

## How It Works

1. **Parallel Processing**: Each CSV file is processed by a separate goroutine
2. **Worker Pool**: A semaphore limits concurrent workers to prevent resource exhaustion
3. **Batch Inserts**: Data is inserted in configurable batches using transactions
4. **SQLite Optimization**: Uses WAL mode, NORMAL synchronous mode, and memory cache
5. **Table Creation**: Automatically creates tables from CSV headers
6. **Name Sanitization**: Cleans table and column names for SQLite compatibility

## Performance Tips

1. **Increase Batch Size**: For large files, use `-batch 5000` or higher
2. **Adjust Workers**: On systems with many cores, try `-workers 16` or more
3. **SSD Storage**: Save the database on SSD for faster write speeds
4. **RAM**: Ensure sufficient RAM for parallel processing of large files

## Example Output

```
Found 3 CSV file(s) to convert
  1. sales_2024.csv
  2. customers.csv
  3. products.csv

Starting conversion with 8 workers (CPU cores available: 8)
✅ Converted customers.csv -> customers (15000 rows)
✅ Converted products.csv -> products (5000 rows)
✅ Converted sales_2024.csv -> sales_2024 (50000 rows)

🎉 Conversion completed!

📊 Conversion Summary:
============================================================
✅ sales_2024.csv -> Table 'sales_2024' (50000 rows)
✅ customers.csv -> Table 'customers' (15000 rows)
✅ products.csv -> Table 'products' (5000 rows)
============================================================
Success: 3 | Failed: 0 | Total rows: 70000
Database saved to: data/mydata.db
```

## CSV Requirements

- First row must contain column headers
- Files should be properly formatted CSV
- Encoding: UTF-8 recommended
- All data is stored as TEXT type in SQLite (you can add type inference if needed)

## Output

- Database is saved in the `data/` folder
- Each CSV file becomes a separate table
- Table names are sanitized from the CSV filename
- Column names are sanitized from CSV headers

## Troubleshooting

### "Failed to open CSV file"
- Check file path and permissions
- Ensure the file exists and is readable

### "Failed to create database directory"
- Check write permissions for the output directory
- Ensure parent directories exist

### Out of Memory
- Reduce batch size: `-batch 500`
- Reduce workers: `-workers 4`
- Process files individually

## License

Part of the FLARE project.
