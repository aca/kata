package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"math/rand"
)

func main() {
	systemdRun()
}

func systemdRun() {
	cmd := exec.Command("systemd-run", append([]string{"--unit", "xd-" + randomAlphaNumeric(), "--user", "--scope"}, os.Args[1:]...)...)

	// Connect stdout and stderr to the terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("rund:", err)
		return
	}

	// Print the output
	// fmt.Println(string(output))
}

func getAllEnvKeys() []string {
	envs := os.Environ()
	keys := make([]string, 0, len(envs))
	for _, env := range envs {
		// Split each "KEY=value" string into key and value
		pair := strings.SplitN(env, "=", 2)
		key := pair[0] // The key is the first part
		keys = append(keys, key)
	}
	return keys
}

func randomAlphaNumeric() string {
	// Define the character set: a-z, A-Z, 0-9
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 3

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a byte slice to store the result
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		// Pick a random character from the charset
		result[i] = charset[rand.Intn(len(charset))]
	}

	// Convert byte slice to string and return
	return string(result)
}
