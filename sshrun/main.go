package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/go-stdx/shellescape"
)

func main() {
	log.Println("start")
	wd, _ := os.Getwd()
	cmdString := shellescape.Quote(wd)
	_ = cmdString
	cmdString = "cd " + shellescape.Quote(wd) + " &&"
	for _, args := range os.Args[2:] {
		cmdString += " " + shellescape.Quote(args)
	}

	cmd := exec.Command("ssh", os.Args[1], "bash", "-eo", "pipefail", "-c", shellescape.Quote(cmdString))
	log.Println("cmdargs :=", cmd.Args)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// cmd.Stdin = os.Stdin
	// err := cmd.Run()
	// if err != nil {
	// 	panic(err)
	// }
}
