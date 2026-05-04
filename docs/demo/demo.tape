Output docs/demo/demo.gif

Set Width 1400
Set Height 600
Set FontSize 14
Set Theme "Dracula"
Set Shell "bash"
Set Padding 10
Set WindowBar Colorful
Set WindowBarSize 28
Set LoopOffset 20%

# ── Hidden setup ─────────────────────────────────────────────
Set TypingSpeed 1ms
Type "bash /home/rolf/work/home/skillsync/demo/setup-demo.sh && clear"
Enter
Sleep 2s

# ── Launch tmux: left pane (Linux) / right pane (Windows) ────
Type "tmux new-session \; split-window -h \; select-pane -t 0"
Enter
Sleep 800ms

# ── LEFT PANE – Linux machine ────────────────────────────────
Type "PS1='user@linux$ ' && cd /tmp/ss-demo/linux-project && clear && echo '===== Linux machine ====='"
Enter
Sleep 400ms
Set TypingSpeed 50ms

Type "ls .claude/skills/"
Enter
Sleep 600ms

Type "cat .claude/skills/tdd/SKILL.md"
Enter
Sleep 1800ms

Type "skillsync push"
Enter
Sleep 1200ms

# ── Switch to RIGHT PANE ─────────────────────────────────────
Set TypingSpeed 1ms
Type "tmux select-pane -t 1"
Enter
Sleep 300ms

# ── RIGHT PANE – Windows machine ─────────────────────────────
Type "PS1='PS> ' && cd /tmp/ss-demo/windows-project && clear && echo '===== Windows machine ====='"
Enter
Sleep 400ms
Set TypingSpeed 50ms

Type "skillsync list"
Enter
Sleep 800ms

Type "skillsync add skills/tdd"
Enter
Sleep 1500ms

Type "ls .claude/skills/"
Enter
Sleep 600ms

Type "cat .claude/skills/tdd/SKILL.md"
Enter
Sleep 2000ms

# ── Switch back to LEFT PANE – Linux edit ────────────────────
Set TypingSpeed 1ms
Type "tmux select-pane -t 0"
Enter
Sleep 300ms
Set TypingSpeed 50ms

Type "echo '' >> .claude/skills/tdd/SKILL.md"
Enter
Sleep 300ms

Type "echo '## Rule' >> .claude/skills/tdd/SKILL.md"
Enter
Sleep 300ms

Type "echo '' >> .claude/skills/tdd/SKILL.md"
Enter
Sleep 300ms

Type "echo 'Commit only when all tests are green.' >> .claude/skills/tdd/SKILL.md"
Enter
Sleep 500ms

Type "cat .claude/skills/tdd/SKILL.md"
Enter
Sleep 1800ms

Type "skillsync push"
Enter
Sleep 1200ms

# ── Switch to RIGHT PANE – Windows pull ──────────────────────
Set TypingSpeed 1ms
Type "tmux select-pane -t 1"
Enter
Sleep 300ms
Set TypingSpeed 50ms

Type "skillsync pull"
Enter
Sleep 1500ms

Type "cat .claude/skills/tdd/SKILL.md"
Enter
Sleep 3000ms
