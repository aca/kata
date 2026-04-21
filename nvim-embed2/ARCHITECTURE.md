## How it works

### Startup sequence

1. **tcell screen init** (`main.go:58-63`) — Creates a virtual terminal screen via tcell. `s.Size()` returns the terminal dimensions. We reserve the bottom 4 rows (3 for message box + 1 separator), giving the rest to neovim.

2. **Neovim child process** (`main.go:73-79`) — `nvim.NewChildProcess` spawns `nvim --embed` as a subprocess. In this mode, neovim does **not** render to a terminal. Instead, its stdin/stdout become a **MessagePack-RPC** channel. The go-client library handles serialization, request/response matching, and notification dispatch over this pipe. Internally it starts a goroutine (`Serve`) that reads msgpack frames from neovim's stdout in a loop.

3. **Register redraw handler** (`main.go:81-83`) — Before any UI events arrive, we register a handler for the `"redraw"` RPC notification method. Neovim sends all UI updates as `redraw` notifications. The go-client's RPC layer deserializes the msgpack payload and calls our function via reflection. The variadic `...[]any` signature matches how neovim batches multiple update events into a single notification — each `[]any` is one event like `["grid_line", [args...], [args...]]`.

4. **UI attach** (`main.go:85-91`) — `AttachUI(width, height, opts)` sends the `nvim_ui_attach` RPC request to neovim. This tells neovim "I am your UI frontend, send me screen updates". The options:
   - `ext_linegrid: true` — Use the newer linegrid protocol (events like `grid_line`, `grid_scroll`) instead of the legacy cell-by-cell protocol. This is more efficient — neovim sends whole line segments at once.
   - `rgb: true` — Report colors as 24-bit RGB values (we currently ignore them but neovim requires one of rgb/term).

   After this call, neovim begins sending `redraw` notifications whenever the screen changes.

### The redraw protocol

Neovim's `redraw` notification contains a batch of updates. Each update is an array where element 0 is the event name and elements 1+ are argument arrays (one per invocation of that event). This batching is why the handler iterates twice: outer loop over updates, inner loop over event instances.

The events we handle:

- **`grid_resize [grid, width, height]`** — Neovim tells us the grid dimensions changed. We reallocate the cell buffer. Grid ID 1 is the default global grid (we ignore the grid ID since we don't use `ext_multigrid`).

- **`grid_line [grid, row, col_start, cells]`** (`main.go:143-173`) — The core rendering event. `cells` is an array of `[text, hl_id, repeat]` tuples:
  - `text`: a single character (as a UTF-8 string)
  - `hl_id` (optional): highlight group ID, carried forward if omitted
  - `repeat` (optional, default 1): how many consecutive cells get this character (used for runs of spaces)
  
  We walk left-to-right from `col_start`, writing characters into our cell buffer.

- **`grid_cursor_goto [grid, row, col]`** — Neovim reports where the cursor is. We store it and pass it to `tcell.ShowCursor()` during render.

- **`grid_clear [grid]`** — Fill the entire grid with spaces. Sent on `:edit`, screen redraws, etc.

- **`grid_scroll [grid, top, bot, left, right, rows, cols]`** (`main.go:186-221`) — A rectangular region scrolls. `rows > 0` means content moves up (new blank lines appear at the bottom of the region); `rows < 0` means content moves down. We copy cells in the appropriate direction and blank the vacated area. This avoids neovim having to resend every line when you scroll.

- **`flush`** — Marks the end of an atomic batch of updates. This is the only place we call `render()`. Rendering on flush (not on every individual event) means we get consistent frames and don't waste time painting intermediate states.

### Rendering (`main.go:224-264`)

Called with `a.mu` held. Writes to the tcell screen buffer (not the real terminal — tcell double-buffers):

1. **Top pane**: Copy our cell buffer to tcell positions `(col, row)` for rows `0..nvHeight-1`.
2. **Separator**: Draw a row of `─` characters at row `nvHeight` with a colored background.
3. **Message box**: Clear the bottom 3 rows with a blue background, center "Hello, World!" vertically and horizontally.
4. **Cursor**: `ShowCursor(col, row)` tells tcell where to place the terminal cursor — this maps directly to neovim's cursor position.
5. `s.Show()` diffs the internal buffer against what's on screen and emits only the necessary terminal escape sequences.

### Input forwarding (`main.go:266-286`)

Runs on the main goroutine. `s.PollEvent()` blocks until tcell has a terminal event.

- **Key events**: Translated to Neovim's input notation (`<CR>`, `<Esc>`, `<C-w>`, etc.) by `tcellKeyToNvim` and sent via `nv.Input(keys)`. This is the `nvim_input` RPC call — it's **non-blocking on neovim's side** (queues input to neovim's event loop), but the go-client still does a synchronous RPC round-trip. The key translation handles runes (regular characters), modifier combos (Alt+x → `<M-x>`), and special keys (arrows, function keys, etc.).

- **Resize events**: When the terminal resizes, we update our dimensions, reallocate the cell buffer, and call `nv.TryResizeUI(w, h)` which sends `nvim_ui_try_resize`. Neovim then reflows its windows and sends new `redraw` events with the updated layout. `s.Sync()` forces tcell to repaint everything since the terminal size changed.

- **Ctrl+Q**: Short-circuits the loop and returns, causing `main` to exit (deferred `s.Fini()` restores the terminal, deferred `nv.Close()` kills the neovim process).

### Threading model

```
Main goroutine          go-client reader          go-client notifier
─────────────          ─────────────────          ──────────────────
inputLoop()            reads msgpack frames       runs notification
  PollEvent()          from nvim stdout,          handlers serially
  Input() ──RPC──►     dispatches replies         in order received
                       to pending calls,          
                       queues notifications ──►   handleRedraw()
                                                    locks a.mu
                                                    updates cells
                                                    on flush: render()
                                                    unlocks a.mu
```

Three goroutines are in play:
1. **Main** — runs the tcell event loop, sends input to neovim via RPC
2. **go-client reader** — reads from neovim's stdout pipe, deserializes msgpack, routes replies to blocked `Call`s and queues notifications
3. **go-client notifier** — dequeues notifications and calls handlers sequentially (ensures ordering)

The mutex `a.mu` protects the shared cell buffer: the notifier goroutine writes to it in `handleRedraw`, and resize events on the main goroutine reallocate it. `render()` is always called with the lock held.

### What we don't handle

- **Highlight attributes** (`hl_attr_define`) — Neovim sends color/bold/italic info per highlight ID. We store `hlID` nowhere and render everything in the default terminal style.
- **Multi-grid** (`ext_multigrid`) — We treat everything as one grid. Neovim windows, floating windows, and the message area all composite onto grid 1.
- **Mouse events** — `nvim_input_mouse` exists but we don't forward tcell mouse events.
- **Wide characters** — CJK characters that occupy 2 cells. Neovim sends an empty string `""` for the second cell; we'd need to handle `tcell.SetContent` with combining characters.
- **Mode changes** (`mode_info_set`, `mode_change`) — cursor shape changes between insert/normal mode.
