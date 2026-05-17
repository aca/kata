package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var hostname = func() string {
	h, _ := os.Hostname()
	return h
}()

func osc7(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return fmt.Sprintf("\x1b]7;file://%s%s\x1b\\", hostname, strings.Join(parts, "/"))
}

type entry struct {
	name  string
	size  int64
	isDir bool
}

type scanEntryMsg struct {
	path string
	e    entry
	ch   <-chan tea.Msg
}

type scanDoneMsg struct {
	path string
}

type model struct {
	cwd       string
	entries   []entry
	cursor    int
	cache     map[string][]entry
	done      map[string]bool
	scanning  bool
	width     int
	height    int
	confirm   string // full path pending delete; empty if no prompt
	statusMsg string
}

func initialModel(start string) model {
	return model{
		cwd:      start,
		cache:    map[string][]entry{},
		done:     map[string]bool{},
		scanning: true,
	}
}

func startScan(path string) tea.Cmd {
	ch := make(chan tea.Msg, 32)
	go func() {
		defer close(ch)
		dirents, err := os.ReadDir(path)
		if err != nil {
			return
		}
		for _, d := range dirents {
			full := filepath.Join(path, d.Name())
			e := entry{name: d.Name(), isDir: d.IsDir()}
			if d.IsDir() {
				e.size = dirSize(full)
			} else {
				if info, err := d.Info(); err == nil {
					e.size = info.Size()
				}
			}
			ch <- scanEntryMsg{path: path, e: e, ch: ch}
		}
	}()
	return readScan(path, ch)
}

func readScan(path string, ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return scanDoneMsg{path: path}
		}
		return msg
	}
}

func insertSorted(es []entry, e entry) []entry {
	i := sort.Search(len(es), func(i int) bool { return es[i].size < e.size })
	es = append(es, entry{})
	copy(es[i+1:], es[i:])
	es[i] = e
	return es
}

func dirSize(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func (m model) Init() tea.Cmd {
	return startScan(m.cwd)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case scanEntryMsg:
		m.cache[msg.path] = insertSorted(m.cache[msg.path], msg.e)
		if msg.path == m.cwd {
			m.entries = m.cache[msg.path]
		}
		return m, readScan(msg.path, msg.ch)

	case scanDoneMsg:
		m.done[msg.path] = true
		if msg.path == m.cwd {
			m.scanning = false
		}
		return m, nil

	case tea.KeyMsg:
		if m.confirm != "" {
			if msg.String() == "enter" {
				target := m.confirm
				m.confirm = ""
				if err := os.RemoveAll(target); err != nil {
					m.statusMsg = "delete failed: " + err.Error()
					return m, nil
				}
				name := filepath.Base(target)
				kept := m.entries[:0]
				for _, e := range m.entries {
					if e.name != name {
						kept = append(kept, e)
					}
				}
				m.entries = kept
				m.cache[m.cwd] = kept
				if m.cursor >= len(m.entries) {
					m.cursor = len(m.entries) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.statusMsg = "deleted " + name
				return m, nil
			}
			m.confirm = ""
			m.statusMsg = "cancelled"
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "d":
			if len(m.entries) == 0 {
				return m, nil
			}
			m.confirm = filepath.Join(m.cwd, m.entries[m.cursor].name)
			return m, nil
		case "j", "down":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "g":
			m.cursor = 0
		case "G":
			if len(m.entries) > 0 {
				m.cursor = len(m.entries) - 1
			}
		case "h", "left":
			parent := filepath.Dir(m.cwd)
			if parent != m.cwd {
				prev := filepath.Base(m.cwd)
				m.cwd = parent
				m.entries = m.cache[parent]
				m.scanning = !m.done[parent]
				m.cursor = 0
				for i, e := range m.entries {
					if e.name == prev {
						m.cursor = i
						break
					}
				}
				if m.scanning && m.entries == nil {
					return m, startScan(parent)
				}
				return m, nil
			}
		case "l", "right", "enter":
			if len(m.entries) == 0 {
				return m, nil
			}
			sel := m.entries[m.cursor]
			full := filepath.Join(m.cwd, sel.name)
			if sel.isDir {
				m.cwd = full
				m.cursor = 0
				m.entries = m.cache[full]
				if m.done[full] {
					m.scanning = false
					return m, nil
				}
				m.scanning = true
				if m.entries == nil {
					return m, startScan(full)
				}
				return m, nil
			}
			if msg.String() == "enter" {
				return m, openCmd(full)
			}
		case "o":
			if len(m.entries) == 0 {
				return m, nil
			}
			sel := m.entries[m.cursor]
			full := filepath.Join(m.cwd, sel.name)
			return m, openCmd(full)
		case "r":
			delete(m.cache, m.cwd)
			delete(m.done, m.cwd)
			m.entries = nil
			m.scanning = true
			return m, startScan(m.cwd)
		}
	}
	return m, nil
}

func openCmd(path string) tea.Cmd {
	return func() tea.Msg {
		_ = exec.Command("mpv", path).Start()
		return nil
	}
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dirStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	cursorStyle = lipgloss.NewStyle().Reverse(true)
)

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func (m model) View() string {
	cwdSeq := osc7(m.cwd)
	header := cwdSeq + headerStyle.Render(fmt.Sprintf("dux: %s", m.cwd))
	if len(m.entries) == 0 {
		state := "(empty)"
		if m.scanning {
			state = "scanning..."
		}
		return header + "\n\n" + state + "\n\nh: up  ctrl+c: quit\n"
	}

	var total int64
	for _, e := range m.entries {
		total += e.size
	}

	// chrome: header line + blank + footer blank + footer line + scroll indicator line = 5
	visible := m.height - 5
	if visible < 1 {
		visible = len(m.entries)
	}
	offset := 0
	if m.cursor >= visible {
		offset = m.cursor - visible + 1
	}
	end := offset + visible
	if end > len(m.entries) {
		end = len(m.entries)
	}

	body := ""
	for i := offset; i < end; i++ {
		e := m.entries[i]
		name := e.name
		if e.isDir {
			name = name + "/"
		}
		var pct float64
		if total > 0 {
			pct = float64(e.size) / float64(total) * 100
		}
		line := fmt.Sprintf(" %10s  %s  %5.1f%%  %s", humanSize(e.size), bar(pct, 20), pct, name)
		if e.isDir {
			line = dirStyle.Render(line)
		}
		if i == m.cursor {
			line = cursorStyle.Render(line)
		}
		body += line + "\n"
	}

	var footer string
	if m.confirm != "" {
		footer = cursorStyle.Render(fmt.Sprintf(" delete %q ? [enter=yes, any other=cancel] ", filepath.Base(m.confirm)))
	} else {
		status := ""
		if m.scanning {
			status = " scanning..."
		} else if m.statusMsg != "" {
			status = " " + m.statusMsg
		}
		scroll := fmt.Sprintf(" [%d-%d/%d]%s", offset+1, end, len(m.entries), status)
		footer = fmt.Sprintf("%s total: %s   [j/k move  l/enter open  h up  o open  d del  r rescan  ^C quit]",
			scroll, humanSize(total))
	}

	return header + "\n" + body + footer + "\n"
}

func bar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	s := "["
	for i := 0; i < width; i++ {
		if i < filled {
			s += "#"
		} else {
			s += " "
		}
	}
	s += "]"
	return s
}

func main() {
	start, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		start = os.Args[1]
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(abs), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
