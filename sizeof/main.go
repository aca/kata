package main

import (
	"fmt"
	"os"
)

func main() {
	stat, err := os.Stat(os.Args[1])
	if err != nil {
		panic(err)
	}
	_ = err
	_ = stat

	fmt.Println(stat.Size())
}
