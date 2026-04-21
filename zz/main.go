package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	
	"github.com/kballard/go-shellquote"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --list\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --interactive\n", os.Args[0])
		os.Exit(1)
	}

	// Handle --list flag
	if os.Args[1] == "--list" {
		listRunningCommands()
		return
	}
	
	// Handle --interactive flag
	if os.Args[1] == "--interactive" {
		runInteractiveTUI()
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

	// Build the full command with all arguments
	fullCmd := []string{command}
	fullCmd = append(fullCmd, args...)
	
	// Use shellquote.Join to properly escape the command
	quotedCmd := shellquote.Join(fullCmd...)

	// Create wrapper script content
	scriptContent := fmt.Sprintf(`#!/bin/bash
# Run the command
%s
EXIT_CODE=$?

# Send notification based on exit code
if [ $EXIT_CODE -eq 0 ]; then
    notify-send 'zz' '%s done'
else
    notify-send -u critical 'zz' '%s failed (exit code: '$EXIT_CODE')'
fi

exit $EXIT_CODE
`, quotedCmd, cmdDisplayEscaped, cmdDisplayEscaped)

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

type unitInfo struct {
	name        string
	status      string
	description string
	startTime   time.Time
	endTime     *time.Time
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

	// Parse units and collect info
	var units []unitInfo
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Remove leading ● if present
		line = strings.TrimPrefix(line, "● ")
		
		// Extract unit name and status from the line
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		unitName := fields[0]
		loadState := fields[1]
		activeState := fields[2]
		subState := fields[3]
		
		// Skip if not loaded or doesn't match our pattern
		if loadState == "not-found" || !strings.HasPrefix(unitName, "zz-") || !strings.HasSuffix(unitName, ".service") {
			continue
		}
		
		// Get properties from the unit
		showCmd := exec.Command("systemctl", "--user", "show", 
			"-p", "Description",
			"-p", "ActiveEnterTimestamp", 
			"-p", "ActiveExitTimestamp",
			"-p", "InactiveEnterTimestamp",
			"-p", "ExecMainExitTimestamp",
			unitName)
		propOutput, err := showCmd.Output()
		if err != nil {
			continue
		}
		
		// Parse properties
		props := make(map[string]string)
		for _, propLine := range strings.Split(string(propOutput), "\n") {
			if idx := strings.Index(propLine, "="); idx > 0 {
				key := propLine[:idx]
				value := propLine[idx+1:]
				props[key] = value
			}
		}
		
		description := strings.TrimSpace(props["Description"])
		
		// Parse timestamps
		var startTime time.Time
		var endTime *time.Time
		
		if activeEnter := props["ActiveEnterTimestamp"]; activeEnter != "" && activeEnter != "n/a" {
			if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", activeEnter); err == nil {
				startTime = t
			}
		}
		
		// For end time, check ExecMainExitTimestamp for completed services
		if activeState == "active" && subState == "exited" {
			// Service completed successfully, check ExecMainExitTimestamp
			if execMainExit := props["ExecMainExitTimestamp"]; execMainExit != "" && execMainExit != "n/a" {
				if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", execMainExit); err == nil {
					endTime = &t
				}
			}
		} else if activeState != "active" {
			// Service failed or stopped, check InactiveEnterTimestamp or ActiveExitTimestamp
			if inactiveEnter := props["InactiveEnterTimestamp"]; inactiveEnter != "" && inactiveEnter != "n/a" {
				if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", inactiveEnter); err == nil {
					endTime = &t
				}
			} else if activeExit := props["ActiveExitTimestamp"]; activeExit != "" && activeExit != "n/a" {
				if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", activeExit); err == nil {
					endTime = &t
				}
			}
		}
		
		// Format the status
		status := fmt.Sprintf("%s/%s", activeState, subState)
		if activeState == "active" && subState == "running" {
			status = "running"
		} else if activeState == "active" && subState == "exited" {
			status = "completed"
		} else if activeState == "failed" {
			status = "failed"
		}
		
		units = append(units, unitInfo{
			name:        unitName,
			status:      status,
			description: description,
			startTime:   startTime,
			endTime:     endTime,
		})
	}
	
	// Sort by start time (oldest first)
	sort.Slice(units, func(i, j int) bool {
		return units[i].startTime.Before(units[j].startTime)
	})
	
	// Display results with header
	fmt.Println("Zz commands:")
	fmt.Printf("%-10s %-10s %-10s %-30s %-12s %s\n", 
		"Started", "Ended", "Duration", "Unit", "Status", "Command")
	fmt.Println(strings.Repeat("-", 120))
	
	for _, unit := range units {
		startStr := "-"
		endStr := "-"
		durationStr := "-"
		
		if !unit.startTime.IsZero() {
			startStr = unit.startTime.Format("15:04:05")
			
			if unit.endTime != nil {
				endStr = unit.endTime.Format("15:04:05")
				duration := unit.endTime.Sub(unit.startTime)
				durationStr = formatDuration(duration)
			} else if unit.status == "running" {
				duration := time.Since(unit.startTime)
				durationStr = formatDuration(duration)
			}
		}
		
		fmt.Printf("%-10s %-10s %-10s %-30s %-12s %s\n", 
			startStr, endStr, durationStr, unit.name, unit.status, unit.description)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}