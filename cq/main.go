package main

import (
	"os"
	"os/exec"
)

var SHELL = "zsh"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cmd := exec.Command(SHELL, "/dev/stdin")
	stdinpipe, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_, err = stdinpipe.Write([]byte("fail 3"))
	must(err)
	must(stdinpipe.Close())
	must(cmd.Run())
}
