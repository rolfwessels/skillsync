# record-demo

Create an animated terminal demo GIF using [VHS](https://github.com/charmbracelet/vhs).

## Prerequisites

Install on first use (requires sudo for the apt repo):

```bash
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg
echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
sudo apt-get update -qq && sudo apt-get install -y vhs ffmpeg ttyd
```

The apt-packaged `ttyd` is likely too old (VHS requires ≥ 1.7.2). Replace it:

```bash
curl -fsSL -o ~/.local/bin/ttyd https://github.com/tsl0922/ttyd/releases/download/1.7.7/ttyd.x86_64
chmod +x ~/.local/bin/ttyd
```

VHS downloads Chromium on first run (~500 MB, cached at `~/.cache/rod/`).

Run from the repo root: `vhs docs/demo/demo.tape`

## Tape file structure

```tape
Output docs/demo/demo.gif

Set Width 1400
Set Height 600
Set FontSize 14
Set Theme "Dracula"
Set Shell "bash"
Set Padding 10
Set WindowBar Colorful    # macOS-style title bar
Set WindowBarSize 28
Set LoopOffset 20%        # skip past setup on loop-back
```

## Two-pane tmux pattern

```tape
# Start tmux with left/right split, focus left pane
Set TypingSpeed 1ms
Type "tmux new-session \; split-window -h \; select-pane -t 0"
Enter
Sleep 800ms

# Left pane setup
Type "PS1='user@linux$ ' && cd /path/to/left && clear && echo '===== Left machine ====='"
Enter
Sleep 400ms
Set TypingSpeed 50ms

# ... left pane commands ...

# Switch to right pane — use select-pane, NOT Ctrl+B arrow (doesn't work in VHS)
Set TypingSpeed 1ms
Type "tmux select-pane -t 1"
Enter
Sleep 300ms

# Right pane setup
Type "PS1='PS> ' && cd /path/to/right && clear && echo '===== Right machine ====='"
Enter
Sleep 400ms
Set TypingSpeed 50ms

# ... right pane commands ...

# Switch back to left pane
Set TypingSpeed 1ms
Type "tmux select-pane -t 0"
Enter
Sleep 300ms
Set TypingSpeed 50ms
```

## Key lessons

- **`Ctrl+B` + `Right` does not work in VHS** for tmux pane switching. Use `tmux select-pane -t N` instead.
- Use `Set TypingSpeed 1ms` for setup/navigation commands so they flash by invisibly, then restore to `50ms` for demo commands.
- `Set LoopOffset 20%` makes the GIF loop back past the initial setup so it looks clean on repeat.
- The `Output` path is relative to where `vhs` is invoked (always run from the repo root).
- Pair the tape with a `setup-demo.sh` that builds the demo environment; call it at the start of the tape with `> /dev/null 2>&1 && clear` to hide setup noise.
- `Set WindowBar Colorful` adds a macOS-style window chrome that makes the GIF look polished.
