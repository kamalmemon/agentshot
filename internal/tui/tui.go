package tui

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/google/uuid"
)

// ANSI color codes to hex (One Dark theme)
var ansiColors = map[int]string{
	0:  "#282c34", // black
	1:  "#e06c75", // red
	2:  "#98c379", // green
	3:  "#e5c07b", // yellow
	4:  "#61afef", // blue
	5:  "#c678dd", // magenta
	6:  "#56b6c2", // cyan
	7:  "#abb2bf", // white
	8:  "#5c6370", // bright black
	9:  "#e06c75", // bright red
	10: "#98c379", // bright green
	11: "#e5c07b", // bright yellow
	12: "#61afef", // bright blue
	13: "#c678dd", // bright magenta
	14: "#56b6c2", // bright cyan
	15: "#ffffff", // bright white
}

const (
	defaultBg = "#282c34"
	defaultFg = "#abb2bf"
)

type cell struct {
	char   rune
	fg     string
	bg     string
	bold   bool
	dim    bool
	italic bool
}

type screen struct {
	cells  [][]cell
	cols   int
	rows   int
	curX   int
	curY   int
	curFg  string
	curBg  string
	bold   bool
	dim    bool
	italic bool
}

func newScreen(cols, rows int) *screen {
	s := &screen{
		cols:  cols,
		rows:  rows,
		cells: make([][]cell, rows),
		curFg: defaultFg,
		curBg: "",
	}
	for i := range s.cells {
		s.cells[i] = make([]cell, cols)
		for j := range s.cells[i] {
			s.cells[i][j] = cell{char: ' ', fg: defaultFg}
		}
	}
	return s
}

func (s *screen) write(r rune) {
	if s.curX >= s.cols {
		s.curX = 0
		s.curY++
	}
	if s.curY >= s.rows {
		// Scroll up
		copy(s.cells, s.cells[1:])
		s.cells[s.rows-1] = make([]cell, s.cols)
		for j := range s.cells[s.rows-1] {
			s.cells[s.rows-1][j] = cell{char: ' ', fg: defaultFg}
		}
		s.curY = s.rows - 1
	}
	if s.curY >= 0 && s.curX >= 0 && s.curY < s.rows && s.curX < s.cols {
		s.cells[s.curY][s.curX] = cell{
			char:   r,
			fg:     s.curFg,
			bg:     s.curBg,
			bold:   s.bold,
			dim:    s.dim,
			italic: s.italic,
		}
	}
	s.curX++
}

func (s *screen) newline() {
	s.curX = 0
	s.curY++
}

func (s *screen) carriageReturn() {
	s.curX = 0
}

func (s *screen) setSGR(params []int) {
	if len(params) == 0 {
		params = []int{0}
	}

	i := 0
	for i < len(params) {
		p := params[i]
		switch {
		case p == 0:
			s.curFg = defaultFg
			s.curBg = ""
			s.bold = false
			s.dim = false
			s.italic = false
		case p == 1:
			s.bold = true
		case p == 2:
			s.dim = true
		case p == 3:
			s.italic = true
		case p == 22:
			s.bold = false
			s.dim = false
		case p == 23:
			s.italic = false
		case p >= 30 && p <= 37:
			s.curFg = ansiColors[p-30]
		case p == 38:
			// Extended foreground color
			if i+1 < len(params) {
				if params[i+1] == 5 && i+2 < len(params) {
					// 256-color
					s.curFg = color256ToHex(params[i+2])
					i += 2
				} else if params[i+1] == 2 && i+4 < len(params) {
					// RGB
					s.curFg = fmt.Sprintf("#%02x%02x%02x", params[i+2], params[i+3], params[i+4])
					i += 4
				}
			}
		case p == 39:
			s.curFg = defaultFg
		case p >= 40 && p <= 47:
			s.curBg = ansiColors[p-40]
		case p == 48:
			// Extended background color
			if i+1 < len(params) {
				if params[i+1] == 5 && i+2 < len(params) {
					s.curBg = color256ToHex(params[i+2])
					i += 2
				} else if params[i+1] == 2 && i+4 < len(params) {
					s.curBg = fmt.Sprintf("#%02x%02x%02x", params[i+2], params[i+3], params[i+4])
					i += 4
				}
			}
		case p == 49:
			s.curBg = ""
		case p >= 90 && p <= 97:
			s.curFg = ansiColors[p-90+8]
		case p >= 100 && p <= 107:
			s.curBg = ansiColors[p-100+8]
		}
		i++
	}
}

func color256ToHex(n int) string {
	if n < 16 {
		return ansiColors[n]
	}
	if n >= 232 {
		// Grayscale
		gray := (n-232)*10 + 8
		return fmt.Sprintf("#%02x%02x%02x", gray, gray, gray)
	}
	// 216 color cube
	n -= 16
	b := n % 6
	g := (n / 6) % 6
	r := n / 36
	return fmt.Sprintf("#%02x%02x%02x", r*51, g*51, b*51)
}

var ansiEscapeRe = regexp.MustCompile(`\x1b\[([0-9;]*)([A-Za-z])`)

func (s *screen) feed(data []byte) {
	text := string(data)
	lastEnd := 0

	for _, match := range ansiEscapeRe.FindAllStringSubmatchIndex(text, -1) {
		// Write text before escape sequence
		for _, r := range text[lastEnd:match[0]] {
			switch r {
			case '\n':
				s.newline()
			case '\r':
				s.carriageReturn()
			case '\t':
				spaces := 8 - (s.curX % 8)
				for i := 0; i < spaces; i++ {
					s.write(' ')
				}
			default:
				if r >= 32 {
					s.write(r)
				}
			}
		}

		// Process escape sequence
		paramsStr := text[match[2]:match[3]]
		cmd := text[match[4]:match[5]]

		switch cmd {
		case "m": // SGR
			params := parseParams(paramsStr)
			s.setSGR(params)
		case "H", "f": // Cursor position
			params := parseParams(paramsStr)
			row, col := 1, 1
			if len(params) >= 1 {
				row = params[0]
			}
			if len(params) >= 2 {
				col = params[1]
			}
			s.curY = row - 1
			s.curX = col - 1
		case "J": // Erase display
			params := parseParams(paramsStr)
			n := 0
			if len(params) > 0 {
				n = params[0]
			}
			switch n {
			case 2, 3: // Clear entire screen
				for i := range s.cells {
					for j := range s.cells[i] {
						s.cells[i][j] = cell{char: ' ', fg: defaultFg}
					}
				}
				s.curX, s.curY = 0, 0
			}
		case "K": // Erase line
			params := parseParams(paramsStr)
			n := 0
			if len(params) > 0 {
				n = params[0]
			}
			switch n {
			case 0: // Clear to end of line
				for x := s.curX; x < s.cols; x++ {
					s.cells[s.curY][x] = cell{char: ' ', fg: defaultFg}
				}
			case 1: // Clear to start of line
				for x := 0; x <= s.curX; x++ {
					s.cells[s.curY][x] = cell{char: ' ', fg: defaultFg}
				}
			case 2: // Clear entire line
				for x := 0; x < s.cols; x++ {
					s.cells[s.curY][x] = cell{char: ' ', fg: defaultFg}
				}
			}
		case "A": // Cursor up
			params := parseParams(paramsStr)
			n := 1
			if len(params) > 0 && params[0] > 0 {
				n = params[0]
			}
			s.curY -= n
			if s.curY < 0 {
				s.curY = 0
			}
		case "B": // Cursor down
			params := parseParams(paramsStr)
			n := 1
			if len(params) > 0 && params[0] > 0 {
				n = params[0]
			}
			s.curY += n
			if s.curY >= s.rows {
				s.curY = s.rows - 1
			}
		case "C": // Cursor forward
			params := parseParams(paramsStr)
			n := 1
			if len(params) > 0 && params[0] > 0 {
				n = params[0]
			}
			s.curX += n
			if s.curX >= s.cols {
				s.curX = s.cols - 1
			}
		case "D": // Cursor back
			params := parseParams(paramsStr)
			n := 1
			if len(params) > 0 && params[0] > 0 {
				n = params[0]
			}
			s.curX -= n
			if s.curX < 0 {
				s.curX = 0
			}
		}

		lastEnd = match[1]
	}

	// Write remaining text
	for _, r := range text[lastEnd:] {
		switch r {
		case '\n':
			s.newline()
		case '\r':
			s.carriageReturn()
		case '\t':
			spaces := 8 - (s.curX % 8)
			for i := 0; i < spaces; i++ {
				s.write(' ')
			}
		default:
			if r >= 32 {
				s.write(r)
			}
		}
	}
}

func parseParams(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	params := make([]int, len(parts))
	for i, p := range parts {
		n, _ := strconv.Atoi(p)
		params[i] = n
	}
	return params
}

// sanitizeFontFamily removes characters that could enable SVG attribute injection
func sanitizeFontFamily(font string) string {
	// Only allow alphanumeric, spaces, hyphens, and commas (for font stacks)
	var result strings.Builder
	for _, r := range font {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == ' ' || r == '-' || r == ',' || r == '_' {
			result.WriteRune(r)
		}
	}
	if result.Len() == 0 {
		return "monospace"
	}
	return result.String()
}

func (s *screen) toSVG(fontSize int, fontFamily string) string {
	charWidth := float64(fontSize) * 0.6
	lineHeight := float64(fontSize) * 1.2
	padding := 20.0

	width := int(float64(s.cols)*charWidth + padding*2)
	height := int(float64(s.rows)*lineHeight + padding*2)

	// Sanitize font-family to prevent SVG attribute injection
	safeFontFamily := sanitizeFontFamily(fontFamily)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
`, width, height, width, height))
	buf.WriteString(fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s"/>
`, defaultBg))
	buf.WriteString(fmt.Sprintf(`<g font-family="%s" font-size="%dpx">
`, safeFontFamily, fontSize))

	for row := 0; row < s.rows; row++ {
		y := padding + float64(row+1)*lineHeight - lineHeight*0.2

		// Group consecutive characters with same style
		col := 0
		for col < s.cols {
			c := s.cells[row][col]
			if c.char == ' ' && c.bg == "" {
				col++
				continue
			}

			// Find run of same-styled characters
			startCol := col
			var text strings.Builder
			text.WriteRune(c.char)
			col++

			for col < s.cols {
				next := s.cells[row][col]
				if next.fg != c.fg || next.bg != c.bg || next.bold != c.bold {
					break
				}
				text.WriteRune(next.char)
				col++
			}

			x := padding + float64(startCol)*charWidth
			textRaw := text.String()
			textStr := strings.TrimRight(textRaw, " ")

			// Background rect (even if the run is all spaces).
			if c.bg != "" {
				bgWidth := float64(len(textRaw)) * charWidth
				buf.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>
`,
					x, y-lineHeight+lineHeight*0.2, bgWidth, lineHeight, c.bg))
			}

			if textStr == "" {
				continue
			}

			// Text
			weight := "normal"
			if c.bold {
				weight = "bold"
			}
			buf.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" fill="%s" font-weight="%s" xml:space="preserve">%s</text>
`,
				x, y, c.fg, weight, html.EscapeString(textStr)))
		}
	}

	buf.WriteString("</g>\n</svg>\n")
	return buf.String()
}

func Run(args []string) int {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	output := fs.String("o", "", "Output file path (default: /tmp/screenshots/<uuid>.svg)")
	cols := fs.Int("cols", 120, "Terminal columns")
	rows := fs.Int("rows", 40, "Terminal rows")
	delay := fs.Duration("delay", 500*time.Millisecond, "Delay after command for TUI apps")
	fontSize := fs.Int("font-size", 14, "Font size in pixels")
	fontFamily := fs.String("font", "monospace", "Font family")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `agentshot tui - Capture terminal output as SVG

Usage:
  agentshot tui [options] <command>
  <command> | agentshot tui [options]

Options:
`)
		fs.PrintDefaults()
		fmt.Fprint(os.Stderr, `
Examples:
  agentshot tui "ls -la --color=always"
  agentshot tui -o - "git status"
  agentshot tui -o output.svg "cat README.md"
  echo "Hello" | agentshot tui -o hello.svg
`)
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		fs.Usage()
		return 1
	}

	// Ensure screenshot directory exists
	screenshotDir := "/tmp/screenshots"
	if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create screenshot directory: %v\n", err)
		return 1
	}

	outputPath := *output
	if outputPath == "" {
		outputPath = filepath.Join(screenshotDir, uuid.New().String()+".svg")
	}

	scr := newScreen(*cols, *rows)

	// Check if we have stdin input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Reading from pipe
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
			return 1
		}
		scr.feed(data)
	} else if fs.NArg() >= 1 {
		// Run command
		command := fs.Arg(0)
		data, err := runInPTY(command, *cols, *rows, *delay)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run command: %v\n", err)
			return 1
		}
		scr.feed(data)
	} else {
		fs.Usage()
		return 1
	}

	svg := scr.toSVG(*fontSize, *fontFamily)

	// Output to stdout if "-" or write to file
	if outputPath == "-" {
		fmt.Print(svg)
	} else {
		if err := os.WriteFile(outputPath, []byte(svg), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save SVG: %v\n", err)
			return 1
		}
		fmt.Println(outputPath)
	}
	return 0
}

func runInPTY(command string, cols, rows int, delay time.Duration) ([]byte, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		fmt.Sprintf("COLUMNS=%d", cols),
		fmt.Sprintf("LINES=%d", rows),
	)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}
	defer ptmx.Close()

	// Read output with timeout
	var output bytes.Buffer
	done := make(chan error, 1)

	go func() {
		reader := bufio.NewReader(ptmx)
		buf := make([]byte, 4096)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}
			if err != nil {
				done <- err
				return
			}
		}
	}()

	// Wait for command or timeout
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-cmdDone:
		// Command finished, wait a bit more for output
		time.Sleep(100 * time.Millisecond)
	case <-time.After(delay + 10*time.Second):
		cmd.Process.Kill()
	}

	// Additional delay for TUI apps to render
	if delay > 0 {
		time.Sleep(delay)
	}

	return output.Bytes(), nil
}
