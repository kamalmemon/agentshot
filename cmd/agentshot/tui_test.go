package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestTUIScreenshot(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	tests := []struct {
		name    string
		args    []string
		checkFn func(t *testing.T, data []byte)
	}{
		{
			name: "basic command",
			args: []string{"tui", "echo hello"},
			checkFn: func(t *testing.T, data []byte) {
				if !strings.Contains(string(data), "<svg") {
					t.Error("Output should be SVG")
				}
				if !strings.Contains(string(data), "hello") {
					t.Error("Output should contain 'hello'")
				}
			},
		},
		{
			name: "ls command",
			args: []string{"tui", "ls -la /tmp"},
			checkFn: func(t *testing.T, data []byte) {
				if !strings.Contains(string(data), "<svg") {
					t.Error("Output should be SVG")
				}
			},
		},
		{
			name: "custom dimensions",
			args: []string{"tui", "-cols", "80", "-rows", "24", "echo test"},
			checkFn: func(t *testing.T, data []byte) {
				if !strings.Contains(string(data), "<svg") {
					t.Error("Output should be SVG")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := "/tmp/test_tui_" + strings.ReplaceAll(tt.name, " ", "_") + ".svg"
			args := append(tt.args[:1], append([]string{"-o", outputPath}, tt.args[1:]...)...)

			cmd := exec.Command("./agentshot_test_bin", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}

			data, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			if tt.checkFn != nil {
				tt.checkFn(t, data)
			}

			os.Remove(outputPath)
		})
	}
}

func TestTUIStdout(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	// Test output to stdout with -o -
	cmd := exec.Command("./agentshot_test_bin", "tui", "-o", "-", "echo stdout_test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "<svg") {
		t.Error("Stdout should contain SVG")
	}

	if !strings.Contains(string(output), "stdout_test") {
		t.Error("Stdout should contain command output")
	}
}

func TestTUIColorOutput(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	outputPath := "/tmp/test_tui_color.svg"
	// Use ANSI escape codes for color
	cmd := exec.Command("./agentshot_test_bin", "tui", "-o", outputPath, "printf '\\033[31mred\\033[0m \\033[32mgreen\\033[0m'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	svgContent := string(data)
	if !strings.Contains(svgContent, "<svg") {
		t.Error("Output should be SVG")
	}

	// Check that it contains color-related fill attributes (indicating colors were processed)
	if !strings.Contains(svgContent, "fill=") {
		t.Error("SVG should contain fill attributes for colors")
	}

	os.Remove(outputPath)
}

func TestTUIPipeInput(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	outputPath := "/tmp/test_tui_pipe.svg"
	cmd := exec.Command("./agentshot_test_bin", "tui", "-o", outputPath)
	cmd.Stdin = strings.NewReader("piped content here")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if !strings.Contains(string(data), "piped content here") {
		t.Error("Output should contain piped content")
	}

	os.Remove(outputPath)
}

func TestTUIAutoOutputPath(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	cmd := exec.Command("./agentshot_test_bin", "tui", "echo auto_path_test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputPath := strings.TrimSpace(string(output))

	// Should be in /tmp/screenshots/ with UUID
	if !strings.HasPrefix(outputPath, "/tmp/screenshots/") {
		t.Fatalf("Expected output in /tmp/screenshots/, got: %s", outputPath)
	}

	if !strings.HasSuffix(outputPath, ".svg") {
		t.Fatalf("Expected .svg extension, got: %s", outputPath)
	}

	os.Remove(outputPath)
}
