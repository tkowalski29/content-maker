package extract

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWithNoArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"extract"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestRunWithHelpCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"extract", "help"}, stdout, stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Commands:") {
		t.Errorf("Expected commands list in help, got: %s", output)
	}
	if !strings.Contains(output, "images") {
		t.Error("Expected 'images' command in help")
	}
	if !strings.Contains(output, "frontmatter") {
		t.Error("Expected 'frontmatter' command in help")
	}
}

func TestRunWithUnknownCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"extract", "unknown"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unknown command, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Unknown command") {
		t.Errorf("Expected 'Unknown command' error, got: %s", errOutput)
	}
}

func TestRunImagesWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"extract", "images"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Usage:") {
		t.Errorf("Expected usage message, got: %s", errOutput)
	}
}

func TestRunImagesWithInvalidFile(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{
		"extract",
		"images",
		"-input=nonexistent.md",
		"-output=test.json",
	}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for nonexistent file, got %d", exitCode)
	}
}

func TestRunFrontmatterWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"extract", "frontmatter"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Usage:") {
		t.Errorf("Expected usage message, got: %s", errOutput)
	}
}

func TestRunFrontmatterWithInvalidFile(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{
		"extract",
		"frontmatter",
		"-input=nonexistent.md",
		"-output=test.json",
	}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for nonexistent file, got %d", exitCode)
	}
}

func TestRunWithDashHFlag(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"extract", "-h"}, stdout, stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message for -h, got: %s", output)
	}
}
