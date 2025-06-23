package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --list\n", os.Args[0])
		os.Exit(1)
	}

	// Handle --list flag
	if os.Args[1] == "--list" {
		listRunningCommands()
		return
	}

	// Get command and arguments
	userCommand := os.Args[1]
	userArgs := os.Args[2:]

	// Create wrapper script
	wrapperScript := createWrapperScript(userCommand, userArgs)

	// Prepare environment variables
	envVars := os.Environ()
	envArgs := make([]string, 0, len(envVars)*2)
	for _, env := range envVars {
		envArgs = append(envArgs, "--setenv", env)
	}

	// Build systemd-run command with unit name
	unitName := fmt.Sprintf("zz-%s-%d", filepath.Base(userCommand), os.Getpid())
	cmdDescription := userCommand
	if len(userArgs) > 0 {
		cmdDescription = fmt.Sprintf("%s %s", userCommand, strings.Join(userArgs, " "))
	}
	systemdArgs := []string{"--user", "--unit", unitName, "--remain-after-exit", "--description", cmdDescription}
	systemdArgs = append(systemdArgs, envArgs...)
	systemdArgs = append(systemdArgs, "--", "/usr/bin/env", "bash", wrapperScript)

	// Execute systemd-run
	cmd := exec.Command("systemd-run", systemdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createWrapperScript(command string, args []string) string {
	// Create command string for display
	cmdDisplay := command
	if len(args) > 0 {
		cmdDisplay = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}

	// Escape single quotes in command display for notification
	cmdDisplayEscaped := strings.ReplaceAll(cmdDisplay, "'", "'\"'\"'")

	// Create wrapper script content using bash's printf %q for proper quoting
	scriptContent := fmt.Sprintf(`#!/bin/bash
# Clean up this script after execution
trap "rm -f %s" EXIT

# Build command with properly quoted arguments
cmd=%s
`, "$0", command)

	// Add each argument using printf %q for proper bash quoting
	for _, arg := range args {
		// Use base64 encoding to safely pass the argument to bash
		encoded := base64.StdEncoding.EncodeToString([]byte(arg))
		scriptContent += fmt.Sprintf(`arg=$(echo %s | base64 -d)
cmd="$cmd $(printf '%%q' "$arg")"
`, encoded)
	}

	scriptContent += fmt.Sprintf(`
# Run the command
eval "$cmd"
EXIT_CODE=$?

# Send notification based on exit code
if [ $EXIT_CODE -eq 0 ]; then
    notify-send 'zz' '%s done'
else
    notify-send -u critical 'zz' '%s failed (exit code: '$EXIT_CODE')'
fi

exit $EXIT_CODE
`, cmdDisplayEscaped, cmdDisplayEscaped)

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "zz-wrapper-*.sh")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating wrapper script: %v\n", err)
		os.Exit(1)
	}

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing wrapper script: %v\n", err)
		os.Exit(1)
	}

	if err := tmpFile.Chmod(0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting wrapper script permissions: %v\n", err)
		os.Exit(1)
	}

	tmpFile.Close()
	return tmpFile.Name()
}

func listRunningCommands() {
	// List all systemd units that match our pattern
	cmd := exec.Command("systemctl", "--user", "list-units", "--all", "--type=service", "--no-legend", "--no-pager", "zz-*.service")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing units: %v\n", err)
		os.Exit(1)
	}

	if len(bytes.TrimSpace(output)) == 0 {
		fmt.Println("No zz commands found")
		return
	}

	// Parse and display the output
	lines := strings.Split(string(output), "\n")
	fmt.Println("Zz commands:")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Extract unit name and status from the line
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			unitName := fields[0]
			loadState := fields[1]
			activeState := fields[2]
			subState := fields[3]
			
			// Skip if not loaded
			if loadState == "not-found" {
				continue
			}
			
			// Get the actual command from the unit
			showCmd := exec.Command("systemctl", "--user", "show", "-p", "Description", unitName)
			descOutput, _ := showCmd.Output()
			description := strings.TrimPrefix(string(descOutput), "Description=")
			description = strings.TrimSpace(description)
			
			// Format the status
			status := fmt.Sprintf("%s/%s", activeState, subState)
			if activeState == "active" && subState == "running" {
				status = "running"
			} else if activeState == "active" && subState == "exited" {
				status = "completed"
			} else if activeState == "failed" {
				status = "failed"
			}
			
			fmt.Printf("  %-30s %-12s %s\n", unitName, status, description)
		}
	}
}