package main

import (
	"fmt"
	"reflect"

	"github.com/SanthoshCheemala/LE-PSI/pkg/matrix"
)

func main() {
	vecType := reflect.TypeOf(matrix.Vector{})
	fmt.Printf("matrix.Vector fields:\n")
	for i := 0; i < vecType.NumField(); i++ {
		field := vecType.Field(i)
		fmt.Printf(" - %s: %s (Exported: %v)\n", field.Name, field.Type, field.PkgPath == "")
	}
}
