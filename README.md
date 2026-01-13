# agentshot

Screenshot tool for AI coding agents. Capture browser pages (PNG) and terminal output (SVG).

## Installation

### 1. Install Go (if needed)

Get Go from https://go.dev/dl/ or use your package manager:

```bash
# Ubuntu/Debian
sudo apt install golang-go

# macOS
brew install go
```

### 2. Install Chrome/Chromium

Chrome/Chromium is required for browser screenshots.

```bash
# Ubuntu/Debian
sudo apt install chromium

# macOS
brew install --cask google-chrome
```

### 3. Install agentshot

```bash
git clone https://github.com/kamalmemon/agentshot.git
cd agentshot
make install
```

## How It Works

```
Browser: URL → Chromium (headless) → Chrome DevTools Protocol → PNG
Terminal: Command → PTY → ANSI codes → Parser → SVG
```

## Claude Code setup

### 1. Allow in Claude Code

Add to `~/.claude/settings.json`:

```json
{
  "permissions": {
    "allow": [
      "Bash(agentshot:*)"
    ]
  }
}
```

### 2. Add to CLAUDE.md

Add to your project's `CLAUDE.md` or `~/.claude/CLAUDE.md`:

```markdown
## Screenshots

Use agentshot to capture screenshots when building UI or debugging:
- After UI changes: `agentshot browser -o /tmp/shot.png http://localhost:3000`
- Terminal output: `agentshot tui -o /tmp/shot.svg "<command>"`
Then read the file to view the result.
```

### 3. (Optional) Add slash command

Create `.claude/commands/screenshot.md`:

```markdown
Take a screenshot of $ARGUMENTS and show me the result.

Use agentshot:
- For URLs: `agentshot browser -o /tmp/shot.png <url>`
- For commands: `agentshot tui -o /tmp/shot.svg "<command>"`

Then read the resulting image file.
```

Usage: `/screenshot https://example.com`

## Usage

### Browser (PNG)

```bash
agentshot browser https://example.com
agentshot browser -o page.png https://example.com
agentshot browser -full https://example.com           # full page
agentshot browser -selector "#main" https://example.com
```

| Flag | Default | Description |
|------|---------|-------------|
| `-o` | auto | Output path |
| `-full` | false | Full scrollable page |
| `-selector` | | CSS selector |
| `-width` | 1280 | Viewport width |
| `-height` | 720 | Viewport height |
| `-wait` | | CSS selector to wait for |
| `-delay` | 0 | Wait after load |
| `-timeout` | 30s | Navigation timeout |

### Terminal (SVG)

```bash
agentshot tui "ls -la --color=always"
agentshot tui -o - "git status"                       # stdout
agentshot tui -cols 80 -rows 24 "htop"
```

| Flag | Default | Description |
|------|---------|-------------|
| `-o` | auto | Output path (`-` for stdout) |
| `-cols` | 120 | Terminal width |
| `-rows` | 40 | Terminal height |
| `-delay` | 500ms | Wait for TUI apps |
| `-font-size` | 14 | Font size |
| `-font` | monospace | Font family |

## License

MIT
