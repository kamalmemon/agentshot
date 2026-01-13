package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowserScreenshot(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "basic screenshot",
			args: []string{"browser", "https://example.com"},
		},
		{
			name: "custom viewport",
			args: []string{"browser", "-width", "800", "-height", "600", "https://example.com"},
		},
		{
			name: "with output path",
			args: []string{"browser", "-o", "/tmp/test_screenshot.png", "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./agentshot_test_bin", tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}

			// Get the output path from stdout
			outputPath := strings.TrimSpace(string(output))
			if outputPath == "" {
				t.Fatal("No output path returned")
			}

			// Verify file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Fatalf("Screenshot file not created: %s", outputPath)
			}

			// Verify it's a valid PNG (check magic bytes)
			data, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read screenshot: %v", err)
			}

			if len(data) < 8 {
				t.Fatal("Screenshot file too small")
			}

			// PNG magic bytes: 137 80 78 71 13 10 26 10
			pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
			for i, b := range pngMagic {
				if data[i] != b {
					t.Fatalf("Invalid PNG magic bytes at position %d", i)
				}
			}

			// Clean up
			os.Remove(outputPath)
		})
	}
}

func TestBrowserScreenshotFullPage(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	outputPath := "/tmp/test_fullpage.png"
	cmd := exec.Command("./agentshot_test_bin", "browser", "-full", "-o", outputPath, "https://example.com")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file exists and is a valid PNG
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read screenshot: %v", err)
	}

	if len(data) < 1000 {
		t.Fatalf("Full page screenshot seems too small: %d bytes", len(data))
	}

	os.Remove(outputPath)
}

func TestBrowserScreenshotSelector(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	outputPath := "/tmp/test_selector.png"
	// Use h1 selector which exists on example.com
	cmd := exec.Command("./agentshot_test_bin", "browser", "-selector", "h1", "-o", outputPath, "https://example.com")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Screenshot file not created: %s", outputPath)
	}

	os.Remove(outputPath)
}

func TestBrowserAutoOutputPath(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "agentshot_test_bin")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	defer os.Remove("agentshot_test_bin")

	cmd := exec.Command("./agentshot_test_bin", "browser", "https://example.com")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputPath := strings.TrimSpace(string(output))

	// Should be in /tmp/screenshots/ with UUID
	if !strings.HasPrefix(outputPath, "/tmp/screenshots/") {
		t.Fatalf("Expected output in /tmp/screenshots/, got: %s", outputPath)
	}

	if !strings.HasSuffix(outputPath, ".png") {
		t.Fatalf("Expected .png extension, got: %s", outputPath)
	}

	// Verify UUID format in filename (36 chars + .png)
	filename := filepath.Base(outputPath)
	if len(filename) != 40 { // UUID (36) + .png (4)
		t.Fatalf("Expected UUID filename, got: %s", filename)
	}

	os.Remove(outputPath)
}
