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
	
	fmt.Println("1.05", price * 1.05)
	fmt.Println("1.030", price * 1.03)
	fmt.Println("1.025", price * 1.025)
	fmt.Println("1.020", price * 1.02)
	fmt.Println("1.010", price * 1.01)
	fmt.Println("----", price * 1)
	fmt.Println("0.99", price * 0.99)
}
