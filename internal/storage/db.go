package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

	cols := ""
	for i, v := range columns {
		if i != len(columns) - 1 {
			cols += v + ", "
		} else {
			cols += v
		}
	}
	
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d", cols, tableName, limit)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	
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

func RetriveData(db *sql.DB,tableName string,columns,mergedColumns []string,limit int) []Transaction {
	existingColumns := GetTableColumns(db, tableName)
	validColumns := make([]string, 0)
	
	for _, col := range columns {
		found := false
		for _, existingCol := range existingColumns {
			if strings.EqualFold(col, existingCol) {
				// Use the exact case from the database to avoid case sensitivity issues
				validColumns = append(validColumns, existingCol)
				found = true
				break
			}
		}
		
		if !found {
			fmt.Printf("Warning: Column '%s' not found in table '%s'. Skipping.\n", col, tableName)
		}
	}
	
	if len(validColumns) == 0 {
		fmt.Printf("Error: No valid columns found in table '%s'.\n", tableName)
		return nil
	}
	cols := ""
	for i, v := range validColumns {
		if i != len(columns) - 1 {
			cols += v + ", "
		} else {
			cols += v
		}
	}
	Query := fmt.Sprintf("select %s from %s limit %s",cols,tableName,fmt.Sprint(limit))
	rows,err := db.Query(Query)

	if err != nil{
		log.Fatal(err)
	}
	defer rows.Close()

	columnsName,err := rows.Columns()
	if err != nil{
		log.Fatal(err)
	}
	values := make([]interface{},len(columnsName))
	valuePtr := make([]interface{},len(columnsName))
	for i := range columnsName{
		valuePtr[i] = &values[i]
	}
	var records []Transaction
	for rows.Next(){
		err := rows.Scan(valuePtr...)
		if err != nil{
			log.Fatal(err)
		}
		trans := Transaction{
			Data:make(map[string]string),
		}

		for i,col := range columnsName{
			var valStr string
			val := values[i]
			switch v := val.(type){
			case []byte:
				valStr = string(v)
			default:
				valStr = fmt.Sprintf("%v",v)
			}
			trans.Data[col] = valStr

		}
		records = append(records, trans)
	}
	return records
}

func GetTableColumns(db *sql.DB, tableName string) []string {
	query := fmt.Sprintf("PRAGMA table_info(%s)",tableName)
	rows,err := db.Query(query)

	if err != nil{
		log.Fatal(err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next(){
		var cid, notnull,pk int
		var name, typename string
		var defaultValue interface{}
		if err := rows.Scan(&cid,&name,&typename,&notnull,&defaultValue,&pk); err != nil{
			log.Fatal(err)
		}
		columns = append(columns, name)
	}

	return columns

}