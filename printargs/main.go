package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	for i, v := range os.Args {
		fmt.Printf("%v: %v\n", i, v)
	}

	log.Println("run")
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
