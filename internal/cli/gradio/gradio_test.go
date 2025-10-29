package gradio

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWithNoArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"gradio"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestRunWithOnlyInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"gradio", "-input=test.json"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestRunWithOnlyGradioURL(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"gradio", "-gradio_url=https://example.com"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestRunWithInvalidFile(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{
		"gradio",
		"-input=nonexistent.json",
		"-gradio_url=https://example.com",
	}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for nonexistent file, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "reading input file") && !strings.Contains(errOutput, "no such file") {
		t.Logf("Stderr output: %s", errOutput)
	}
}

func TestRunWithEmptyArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRunWithInvalidFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"gradio", "-invalid-flag"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid flag, got %d", exitCode)
	}
}
