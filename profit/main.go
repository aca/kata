package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/xtdlib/rat"
)

func main() {
	if len(os.Args) == 3 {
		a := rat.Rat(os.Args[1])
		b := rat.Rat(os.Args[2])
		fmt.Printf("%v%%\n", (b.Sub(a)).Quo(a).Mul(100).SetPrecision(2).DecimalString())
		return
	}

	price, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		panic(err)
	}

	fmt.Println("1.05", price*1.05)
	fmt.Println("1.030", price*1.03)
	fmt.Println("1.025", price*1.025)
	fmt.Println("1.020", price*1.02)
	fmt.Println("1.010", price*1.01)
	fmt.Println("----", price*1)
	fmt.Println("0.99", price*0.99)
}
