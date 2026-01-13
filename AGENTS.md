# agentshot

Screenshot tool for capturing browser pages and terminal output.

## When to Use

Use this tool whenever you need to:
- **Verify UI changes** - after modifying frontend code, take a screenshot to confirm the result
- **Debug web apps** - capture the current state of a page to see what's rendering
- **Document terminal output** - capture colorful CLI output, test results, or build logs
- **Check localhost apps** - screenshot `http://localhost:*` to verify your work

## Commands

```bash
# Browser screenshot (PNG)
agentshot browser -o /tmp/shot.png <url>
agentshot browser -o /tmp/shot.png http://localhost:3000

# Full page screenshot
agentshot browser -full -o /tmp/shot.png <url>

# Terminal screenshot (SVG)
agentshot tui -o /tmp/shot.svg "<command>"
agentshot tui -o /tmp/shot.svg "npm test --color=always"
```

## After Capturing

Read the output file to view the screenshot:
```
Read /tmp/shot.png
Read /tmp/shot.svg
```

## Tips

- Use `--color=always` for colored terminal output
- Use `-full` for scrollable pages
- Use `-delay 2s` for pages that need time to load
- Write to `/tmp/` to avoid permission issues
