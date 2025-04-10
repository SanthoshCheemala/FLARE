package utils

import (
	"fmt"
	"os"
)

/* EnsureDataDirectory creates the data directory if it doesn't exist */
func EnsureDataDirectory() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.MkdirAll("data", 0755); err != nil {
			fmt.Printf("Error creating data directory: %v\n", err)
			os.Exit(1)
		}
	}
}

/* GetColumnIndex finds the index of a column in the columns slice
   Returns 0 if the column is not found */
func GetColumnIndex(columns []string, colName string) int {
	for i, col := range columns {
		if col == colName {
			return i
		}
	}
	return 0
}
