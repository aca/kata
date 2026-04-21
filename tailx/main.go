package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	logger := log.New(os.Stdout, "X: ", 0)
	logger.Println("Hello, World!")

	arg := "main.go"
	_ = arg
	logger.Writer().Write([]byte("hello world"))
	cmd := exec.Command("tail", "-f", arg)
	cmd.
}
