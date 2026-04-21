# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o nvim-embed .
./nvim-embed        # Ctrl+Q to quit
```

Logs go to `/tmp/nvim-embed.log`.

## What This Is

A Go TUI that embeds Neovim via `nvim --embed`. The terminal is split into two panes: top is a fully functional Neovim editor (rendered via the UI attach protocol), bottom is a message box.

## Architecture

Single-file app (`main.go`) with three concurrent goroutines:

- **Main goroutine**: tcell event loop (`inputLoop`) — polls terminal events, translates keys to Neovim notation, forwards via `nvim_input` RPC
- **go-client reader**: reads MessagePack-RPC frames from neovim's stdout pipe
- **go-client notifier**: calls `handleRedraw` for `redraw` notifications, which updates the cell buffer and renders on `flush` events

Neovim communicates via the **ext_linegrid** protocol: `grid_line`, `grid_scroll`, `grid_cursor_goto`, `grid_clear`, `grid_resize`, `flush`. The app maintains a 2D cell buffer that mirrors neovim's grid state, then copies it to tcell's screen buffer on each flush.

See `ARCHITECTURE.md` for detailed protocol documentation.

## Key Dependencies

- `github.com/neovim/go-client/nvim` — MessagePack-RPC client for neovim's `--embed` mode
- `github.com/gdamore/tcell/v2` — terminal rendering and input handling
