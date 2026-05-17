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

	fmt.Println(stat.Size())
}
