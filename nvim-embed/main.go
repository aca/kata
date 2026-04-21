package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/neovim/go-client/nvim"
)

type cell struct {
	ch rune
}

type app struct {
	screen tcell.Screen
	nv     *nvim.Nvim
	mu     sync.Mutex
	width  int
	height int
	cells  [][]cell
	curRow int
	curCol int
}

func (a *app) initGrid(w, h int) {
	a.cells = make([][]cell, h)
	for i := range a.cells {
		a.cells[i] = make([]cell, w)
		for j := range a.cells[i] {
			a.cells[i][j] = cell{ch: ' '}
		}
	}
}

func main() {
	logFile, _ := os.Create("/tmp/nvim-embed.log")
	log.SetOutput(logFile)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}
	defer s.Fini()

	w, h := s.Size()

	a := &app{
		screen: s,
		width:  w,
		height: h,
	}
	a.initGrid(w, h)

	a.nv, err = nvim.NewChildProcess(
		nvim.ChildProcessArgs("--embed"),
		nvim.ChildProcessLogf(log.Printf),
		nvim.ChildProcessServe(false),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer a.nv.Close()

	a.nv.RegisterHandler("redraw", func(updates ...[]any) {
		a.handleRedraw(updates...)
	})

	go func() {
		if err := a.nv.Serve(); err != nil {
			log.Println("nvim exited:", err)
		}
		s.Fini()
		os.Exit(0)
	}()

	opts := map[string]any{
		"ext_linegrid": true,
		"rgb":          true,
	}
	if err := a.nv.AttachUI(w, h, opts); err != nil {
		log.Fatal("ui attach:", err)
	}

	// Top: editor buffer, Bottom: terminal running claude
	a.nv.Command("botright split | terminal claude")
	a.nv.Command("startinsert")

	a.inputLoop()
}

func (a *app) handleRedraw(updates ...[]any) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, update := range updates {
		if len(update) == 0 {
			continue
		}
		name, ok := update[0].(string)
		if !ok {
			continue
		}
		for _, ev := range update[1:] {
			args, ok := ev.([]any)
			if !ok {
				continue
			}
			switch name {
			case "grid_resize":
				if len(args) >= 3 {
					w, _ := toInt(args[1])
					h, _ := toInt(args[2])
					if w > 0 && h > 0 {
						a.width = w
						a.height = h
						a.initGrid(w, h)
					}
				}
			case "grid_line":
				a.gridLine(args)
			case "grid_cursor_goto":
				if len(args) >= 3 {
					a.curRow, _ = toInt(args[1])
					a.curCol, _ = toInt(args[2])
				}
			case "grid_clear":
				for i := range a.cells {
					for j := range a.cells[i] {
						a.cells[i][j] = cell{ch: ' '}
					}
				}
			case "grid_scroll":
				a.gridScroll(args)
			case "flush":
				a.render()
			}
		}
	}
}

func (a *app) gridLine(args []any) {
	if len(args) < 4 {
		return
	}
	row, _ := toInt(args[1])
	col, _ := toInt(args[2])
	cells, ok := args[3].([]any)
	if !ok {
		return
	}
	for _, c := range cells {
		arr, ok := c.([]any)
		if !ok || len(arr) == 0 {
			continue
		}
		text, _ := arr[0].(string)
		repeat := 1
		if len(arr) >= 3 {
			repeat, _ = toInt(arr[2])
		}
		r := []rune(text)
		ch := ' '
		if len(r) > 0 {
			ch = r[0]
		}
		for i := 0; i < repeat; i++ {
			if row >= 0 && row < len(a.cells) && col >= 0 && col < len(a.cells[row]) {
				a.cells[row][col] = cell{ch: ch}
			}
			col++
		}
	}
}

func (a *app) gridScroll(args []any) {
	if len(args) < 7 {
		return
	}
	top, _ := toInt(args[1])
	bot, _ := toInt(args[2])
	left, _ := toInt(args[3])
	right, _ := toInt(args[4])
	rows, _ := toInt(args[5])

	if rows > 0 {
		for r := top; r < bot-rows; r++ {
			for c := left; c < right; c++ {
				if r+rows < len(a.cells) {
					a.cells[r][c] = a.cells[r+rows][c]
				}
			}
		}
		for r := bot - rows; r < bot; r++ {
			for c := left; c < right; c++ {
				if r >= 0 && r < len(a.cells) {
					a.cells[r][c] = cell{ch: ' '}
				}
			}
		}
	} else if rows < 0 {
		for r := bot - 1; r >= top-rows; r-- {
			for c := left; c < right; c++ {
				if r+rows >= 0 && r < len(a.cells) {
					a.cells[r][c] = a.cells[r+rows][c]
				}
			}
		}
		for r := top; r < top-rows; r++ {
			for c := left; c < right; c++ {
				if r >= 0 && r < len(a.cells) {
					a.cells[r][c] = cell{ch: ' '}
				}
			}
		}
	}
}

func (a *app) render() {
	s := a.screen
	style := tcell.StyleDefault

	for r := 0; r < len(a.cells); r++ {
		for c := 0; c < len(a.cells[r]); c++ {
			s.SetContent(c, r, a.cells[r][c].ch, nil, style)
		}
	}

	s.ShowCursor(a.curCol, a.curRow)
	s.Show()
}

func (a *app) inputLoop() {
	for {
		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ {
				return
			}
			key := tcellKeyToNvim(ev)
			if key != "" {
				if _, err := a.nv.Input(key); err != nil {
					log.Println("input:", err)
				}
			}
		case *tcell.EventResize:
			w, h := ev.Size()
			a.mu.Lock()
			a.width = w
			a.height = h
			a.initGrid(w, h)
			a.mu.Unlock()
			a.nv.TryResizeUI(w, h)
			a.screen.Sync()
		}
	}
}

func tcellKeyToNvim(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyRune:
		r := ev.Rune()
		if ev.Modifiers()&tcell.ModAlt != 0 {
			return fmt.Sprintf("<M-%c>", r)
		}
		return string(r)
	case tcell.KeyEnter:
		return "<CR>"
	case tcell.KeyEscape:
		return "<Esc>"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "<BS>"
	case tcell.KeyTab:
		return "<Tab>"
	case tcell.KeyUp:
		return "<Up>"
	case tcell.KeyDown:
		return "<Down>"
	case tcell.KeyLeft:
		return "<Left>"
	case tcell.KeyRight:
		return "<Right>"
	case tcell.KeyHome:
		return "<Home>"
	case tcell.KeyEnd:
		return "<End>"
	case tcell.KeyPgUp:
		return "<PageUp>"
	case tcell.KeyPgDn:
		return "<PageDown>"
	case tcell.KeyDelete:
		return "<Del>"
	case tcell.KeyInsert:
		return "<Insert>"
	case tcell.KeyCtrlA:
		return "<C-a>"
	case tcell.KeyCtrlB:
		return "<C-b>"
	case tcell.KeyCtrlC:
		return "<C-c>"
	case tcell.KeyCtrlD:
		return "<C-d>"
	case tcell.KeyCtrlE:
		return "<C-e>"
	case tcell.KeyCtrlF:
		return "<C-f>"
	case tcell.KeyCtrlG:
		return "<C-g>"
	case tcell.KeyCtrlK:
		return "<C-k>"
	case tcell.KeyCtrlL:
		return "<C-l>"
	case tcell.KeyCtrlN:
		return "<C-n>"
	case tcell.KeyCtrlO:
		return "<C-o>"
	case tcell.KeyCtrlP:
		return "<C-p>"
	case tcell.KeyCtrlR:
		return "<C-r>"
	case tcell.KeyCtrlS:
		return "<C-s>"
	case tcell.KeyCtrlT:
		return "<C-t>"
	case tcell.KeyCtrlU:
		return "<C-u>"
	case tcell.KeyCtrlV:
		return "<C-v>"
	case tcell.KeyCtrlW:
		return "<C-w>"
	case tcell.KeyCtrlX:
		return "<C-x>"
	case tcell.KeyCtrlY:
		return "<C-y>"
	case tcell.KeyCtrlZ:
		return "<C-z>"
	case tcell.KeyF1:
		return "<F1>"
	case tcell.KeyF2:
		return "<F2>"
	case tcell.KeyF3:
		return "<F3>"
	case tcell.KeyF4:
		return "<F4>"
	case tcell.KeyF5:
		return "<F5>"
	case tcell.KeyF6:
		return "<F6>"
	case tcell.KeyF7:
		return "<F7>"
	case tcell.KeyF8:
		return "<F8>"
	case tcell.KeyF9:
		return "<F9>"
	case tcell.KeyF10:
		return "<F10>"
	case tcell.KeyF11:
		return "<F11>"
	case tcell.KeyF12:
		return "<F12>"
	default:
		return ""
	}
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int64:
		return int(n), true
	case uint64:
		return int(n), true
	case int:
		return n, true
	case float64:
		return int(n), true
	}
	return 0, false
}
