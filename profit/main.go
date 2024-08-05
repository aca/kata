package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	price, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("1.01", price * 1.01)
	fmt.Println("1.00", price * 1)
	fmt.Println("0.99", price * 0.99)
}
