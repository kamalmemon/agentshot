package browser

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

func Run(args []string) int {
	fs := flag.NewFlagSet("browser", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	output := fs.String("o", "", "Output file path (default: /tmp/screenshots/<uuid>.png)")
	width := fs.Int("width", 1280, "Viewport width in pixels")
	height := fs.Int("height", 720, "Viewport height in pixels")
	timeout := fs.Duration("timeout", 30*time.Second, "Navigation timeout")
	fullPage := fs.Bool("full", false, "Capture full scrollable page")
	selector := fs.String("selector", "", "CSS selector for element screenshot")
	waitFor := fs.String("wait", "", "CSS selector to wait for before screenshot")
	waitDelay := fs.Duration("delay", 0, "Additional delay after page load")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `agentshot browser - Capture web page screenshot

Usage:
  agentshot browser [options] <url>

Options:
`)
		fs.PrintDefaults()
		fmt.Fprint(os.Stderr, `
Examples:
  agentshot browser https://example.com
  agentshot browser -o page.png https://example.com
  agentshot browser -full https://example.com
  agentshot browser -selector "#header" https://example.com
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

	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}
	url := fs.Arg(0)

	// Ensure screenshot directory exists
	screenshotDir := "/tmp/screenshots"
	if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create screenshot directory: %v\n", err)
		return 1
	}

	// Determine output path
	outputPath := *output
	if outputPath == "" {
		outputPath = filepath.Join(screenshotDir, uuid.New().String()+".png")
	}

	// Create browser context with Chrome/Chromium
	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dbus", true),
	)
	if chromePath := resolveChromePath(); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, *timeout)
	defer cancel()

	// Build actions
	var actions []chromedp.Action

	actions = append(actions, chromedp.EmulateViewport(int64(*width), int64(*height)))
	actions = append(actions, chromedp.Navigate(url))
	actions = append(actions, chromedp.WaitReady("body"))

	if *waitFor != "" {
		actions = append(actions, chromedp.WaitVisible(*waitFor))
	}

	if *waitDelay > 0 {
		actions = append(actions, chromedp.Sleep(*waitDelay))
	}

	// Take screenshot
	var buf []byte
	if *selector != "" {
		actions = append(actions,
			chromedp.WaitVisible(*selector, chromedp.ByQuery),
			chromedp.Screenshot(*selector, &buf, chromedp.ByQuery),
		)
	} else if *fullPage {
		actions = append(actions, chromedp.FullScreenshot(&buf, 100))
	} else {
		actions = append(actions, chromedp.CaptureScreenshot(&buf))
	}

	if err := chromedp.Run(ctx, actions...); err != nil {
		fmt.Fprintf(os.Stderr, "Screenshot failed (install Chrome/Chromium or set CHROME_BIN): %v\n", err)
		return 1
	}

	if err := os.WriteFile(outputPath, buf, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save screenshot: %v\n", err)
		return 1
	}

	fmt.Println(outputPath)
	return 0
}

func resolveChromePath() string {
	if envPath := strings.TrimSpace(os.Getenv("CHROME_BIN")); envPath != "" {
		if fileExists(envPath) {
			return envPath
		}
	}

	candidates := []string{
		"/opt/google/chrome/chrome",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/bin/google-chrome",
		"/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/bin/chromium",
		"/bin/chromium-browser",
		"/snap/bin/chromium",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	}
	for _, path := range candidates {
		if fileExists(path) {
			return path
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
