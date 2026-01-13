// Package main provides agentshot - screenshot tools for coding agents.
//
// Usage:
//
//	agentshot browser [options] <url>    - Capture web page screenshot
//	agentshot tui [options] <command>    - Capture terminal output as SVG
//
// Examples:
//
//	agentshot browser -o page.png https://example.com
//	agentshot tui -o listing.svg "ls -la --color=always"
package main

import (
	"fmt"
	"os"

	"agentshot/internal/browser"
	"agentshot/internal/tui"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "browser", "b":
		os.Exit(browser.Run(os.Args[2:]))
	case "tui", "t":
		os.Exit(tui.Run(os.Args[2:]))
	case "version", "-v", "--version":
		fmt.Printf("agentshot v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`agentshot - Screenshot tool for coding agents

Usage:
  agentshot <command> [options]

Commands:
  browser, b    Capture web page screenshot (PNG)
  tui, t        Capture terminal output (SVG)
  version       Show version
  help          Show this help

Browser Examples:
  agentshot browser https://example.com
  agentshot browser -o page.png https://example.com
  agentshot browser -full https://example.com
  agentshot browser -selector "#main" https://example.com

TUI Examples:
  agentshot tui "ls -la --color=always"
  agentshot tui -o - "git status"
  agentshot tui -delay 2s "htop"
  echo "Hello" | agentshot tui -o hello.svg

For command-specific help:
  agentshot browser -help
  agentshot tui -help
`)
}
