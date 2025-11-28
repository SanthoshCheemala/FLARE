# Quick Start Guide: CSV to SQLite Converter

## 🚀 Quick Test

Run the example to see it in action:

```bash
cd /Users/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/FLARE
go run examples/csv_converter_example.go
```

This will:
1. Create sample CSV files in the `data/` folder
2. Convert them to SQLite databases
3. Show you the conversion results

## 📝 Using the Command-Line Tool

### Convert specific CSV files:
```bash
go run cmd/csv_converter/main.go -csv file1.csv,file2.csv -output data/mydb.db
```

### Convert all CSV files in a directory:
```bash
go run cmd/csv_converter/main.go -dir /path/to/csv/folder -output data/mydb.db
```

### With custom settings (8 workers, 5000 rows per batch):
```bash
go run cmd/csv_converter/main.go -dir ./csvfiles -workers 8 -batch 5000 -output data/mydb.db
```

## 💻 Using in Your Code

```go
package main

import (
    "log"
    "github.com/SanthoshCheemala/FLARE/utils"
)

func main() {
    // Simple one-liner
    csvFiles := []string{"sales.csv", "products.csv"}
    err := utils.ConvertCSVToSQLite(csvFiles, "data/output.db")
    if err != nil {
        log.Fatal(err)
    }
}
```

## 🎯 Features

✅ **Parallel Processing** - Uses all CPU cores by default  
✅ **Fast Batch Inserts** - Inserts 1000 rows per transaction  
✅ **Auto Table Creation** - Creates tables from CSV headers  
✅ **Name Sanitization** - Cleans table/column names automatically  
✅ **Progress Tracking** - Shows conversion progress in real-time  
✅ **Error Handling** - Continues on errors, reports at the end  

## 🔧 Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| MaxWorkers | CPU cores | Number of concurrent workers |
| BatchSize | 1000 | Rows per transaction |
| CreateTables | true | Auto-create tables |
| DropExisting | true | Drop tables before creating |

## 📂 Output Location

All databases are saved in the `data/` folder by default.

## 🐛 Troubleshooting

**Error: "No such file or directory"**
- Check your CSV file paths
- Use absolute paths if relative paths don't work

**Out of memory**
- Reduce workers: `-workers 2`
- Reduce batch size: `-batch 500`

**Slow conversion**
- Increase batch size: `-batch 5000`
- Use SSD for database storage

## 📚 More Information

See [CSV_TO_SQLITE.md](./CSV_TO_SQLITE.md) for complete documentation.
