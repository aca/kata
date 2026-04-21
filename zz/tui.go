package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tuiModel struct {
	units        []unitInfo
	selectedIdx  int
	showDetails  bool
	detailOutput string
	width        int
	height       int
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m tuiModel) Init() tea.Cmd {
	return tickCmd()
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.units)-1 {
				m.selectedIdx++
			}
		case "enter":
			m.showDetails = !m.showDetails
			if m.showDetails && m.selectedIdx < len(m.units) {
				// Get detailed status and logs
				unit := m.units[m.selectedIdx]
				m.detailOutput = getUnitDetails(unit.name)
			}
		case "esc":
			m.showDetails = false
		case "r":
			// Refresh units
			m.units = getUnits()
		}
	
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	
	case tickMsg:
		// Update units every second
		m.units = getUnits()
		if m.showDetails && m.selectedIdx < len(m.units) {
			unit := m.units[m.selectedIdx]
			m.detailOutput = getUnitDetails(unit.name)
		}
		return m, tickCmd()
	}
	
	return m, nil
}

func (m tuiModel) View() string {
	if m.showDetails {
		return m.renderDetails()
	}
	return m.renderTable()
}

func (m tuiModel) renderTable() string {
	var s strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	s.WriteString(headerStyle.Render("Zz Commands (Interactive Mode)") + "\n\n")
	
	// Column headers
	headers := fmt.Sprintf("%-10s %-10s %-10s %-30s %-12s %s",
		"Started", "Ended", "Duration", "Unit", "Status", "Command")
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(headers) + "\n")
	s.WriteString(strings.Repeat("─", 120) + "\n")
	
	// Rows
	for i, unit := range m.units {
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
		
		row := fmt.Sprintf("%-10s %-10s %-10s %-30s %-12s %s",
			startStr, endStr, durationStr, unit.name, unit.status, unit.description)
		
		// Highlight selected row
		if i == m.selectedIdx {
			selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("238"))
			s.WriteString(selectedStyle.Render(row) + "\n")
		} else {
			// Color based on status
			var style lipgloss.Style
			switch unit.status {
			case "running":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			case "failed":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			case "completed":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			default:
				style = lipgloss.NewStyle()
			}
			s.WriteString(style.Render(row) + "\n")
		}
	}
	
	// Help
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(helpStyle.Render("↑/k: up  ↓/j: down  Enter: details  r: refresh  q: quit"))
	
	return s.String()
}

func (m tuiModel) renderDetails() string {
	if m.selectedIdx >= len(m.units) {
		return "No unit selected"
	}
	
	unit := m.units[m.selectedIdx]
	
	var s strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	s.WriteString(headerStyle.Render(fmt.Sprintf("Details: %s", unit.name)) + "\n\n")
	
	// Basic info
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	s.WriteString(infoStyle.Render("Command: ") + unit.description + "\n")
	s.WriteString(infoStyle.Render("Status: ") + unit.status + "\n")
	
	if !unit.startTime.IsZero() {
		s.WriteString(infoStyle.Render("Started: ") + unit.startTime.Format("2006-01-02 15:04:05") + "\n")
		if unit.endTime != nil {
			s.WriteString(infoStyle.Render("Ended: ") + unit.endTime.Format("2006-01-02 15:04:05") + "\n")
			duration := unit.endTime.Sub(unit.startTime)
			s.WriteString(infoStyle.Render("Duration: ") + formatDuration(duration) + "\n")
		}
	}
	
	s.WriteString("\n" + strings.Repeat("─", 80) + "\n\n")
	
	// Detailed output
	s.WriteString(m.detailOutput)
	
	// Help
	s.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(helpStyle.Render("Esc/Enter: back  q: quit"))
	
	return s.String()
}

func getUnitDetails(unitName string) string {
	var output strings.Builder
	
	// Get full status
	statusCmd := exec.Command("systemctl", "--user", "status", "--no-pager", unitName)
	statusOut, _ := statusCmd.Output()
	output.WriteString("=== Status ===\n")
	output.WriteString(string(statusOut))
	output.WriteString("\n")
	
	// Get logs
	logsCmd := exec.Command("journalctl", "--user", "-u", unitName, "--no-pager", "-n", "50")
	logsOut, _ := logsCmd.Output()
	output.WriteString("\n=== Logs (last 50 lines) ===\n")
	output.WriteString(string(logsOut))
	
	return output.String()
}

func getUnits() []unitInfo {
	// Use existing listRunningCommands logic but return the units
	cmd := exec.Command("systemctl", "--user", "list-units", "--all", "--type=service", "--no-legend", "--no-pager", "zz-*.service")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	
	if len(bytes.TrimSpace(output)) == 0 {
		return nil
	}
	
	var units []unitInfo
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Remove leading ● if present
		line = strings.TrimPrefix(line, "● ")
		
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
	
	return units
}

func runInteractiveTUI() {
	units := getUnits()
	
	p := tea.NewProgram(tuiModel{
		units:       units,
		selectedIdx: 0,
	}, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
	}
}