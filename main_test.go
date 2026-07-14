package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func init() {
	startServer = func(markdownFile, port string, stdout io.Writer) error {
		return nil
	}
}

func runCLI(args ...string) (string, error) {
	fullArgs := append([]string{"gophern"}, args...)
	var stdout, stderr bytes.Buffer
	err := run(fullArgs, &stdout, &stderr)
	output := stdout.String() + stderr.String()
	return output, err
}

func TestCLIUsage(t *testing.T) {
	output, err := runCLI()
	if err == nil {
		t.Fatal("expected error exit code when running with no arguments, got nil")
	}
	if !strings.Contains(output, "Usage: gophern <command>") {
		t.Errorf("expected usage output, got: %s", output)
	}
}

func TestCLIServeNoFile(t *testing.T) {
	output, err := runCLI("serve")
	if err == nil {
		t.Fatal("expected error exit code when serving without file, got nil")
	}
	if !strings.Contains(output, "Error: Markdown file path required for serve command") {
		t.Errorf("expected error message for missing markdown file, got: %s", output)
	}
}

func TestCLIServeSuccess(t *testing.T) {
	output, err := runCLI("serve", "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v (output: %s)", err, output)
	}
	if !strings.Contains(output, "Serving test.md on port 8080...") {
		t.Errorf("expected serving message, got: %s", output)
	}
}

func TestCLIServeCustomPort(t *testing.T) {
	output, err := runCLI("serve", "-port", "9090", "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v (output: %s)", err, output)
	}
	if !strings.Contains(output, "Serving test.md on port 9090...") {
		t.Errorf("expected serving message with custom port, got: %s", output)
	}
}

func TestCLIExportNoFile(t *testing.T) {
	output, err := runCLI("export")
	if err == nil {
		t.Fatal("expected error exit code when exporting without file, got nil")
	}
	if !strings.Contains(output, "Error: Markdown file path required for export command") {
		t.Errorf("expected error message for missing markdown file, got: %s", output)
	}
}

func TestCLIExportSuccess(t *testing.T) {
	output, err := runCLI("export", "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v (output: %s)", err, output)
	}
	if !strings.Contains(output, "Exporting test.md to presentation.html...") {
		t.Errorf("expected export message, got: %s", output)
	}
}

func TestCLIExportCustomOutput(t *testing.T) {
	output, err := runCLI("export", "-o", "out.html", "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v (output: %s)", err, output)
	}
	if !strings.Contains(output, "Exporting test.md to out.html...") {
		t.Errorf("expected export message with custom output path, got: %s", output)
	}
}
