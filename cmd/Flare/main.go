package main

import (
	"flag"
	"fmt"
	"strings"
)

func main(){
	cols := flag.String("columns","type,amont","which columns to encrypt")
	merged_cols := flag.String("columns-merge","","which columns to merge for encryption")
	encrypt := flag.Bool("encrypt",false,"whether to encrypt or not")
	decrypt := flag.Bool("decrypt",false,"whether to decrypt or not")
	limit := flag.Int("LIMIT",100,"no of rows to encrypt or decrypt from begining")

	flag.Parse()
	columns := strings.Split(*cols,",")
	merged_columns := strings.Split(*merged_cols,",")
	for _,v := range columns{
		fmt.Println(v)
	}
	for _,v := range merged_columns{
		fmt.Println(v)
	}
	fmt.Println(*encrypt,*decrypt,*limit)
}