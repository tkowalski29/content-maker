package brief

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWithNoArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief"}, stdout, stderr)

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

	exitCode := Run([]string{"brief", "help"}, stdout, stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Commands:") {
		t.Errorf("Expected commands list in help, got: %s", output)
	}

	// Check for all subcommands
	expectedCommands := []string{"analyze", "fetch", "serp", "suggest"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("Expected '%s' command in help output", cmd)
		}
	}
}

func TestRunWithUnknownCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "unknown"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unknown command, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Unknown command") {
		t.Errorf("Expected 'Unknown command' error, got: %s", errOutput)
	}
}

func TestRunAnalyzeWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "analyze"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "required") {
		t.Errorf("Expected 'required' error message, got: %s", errOutput)
	}
}

func TestRunFetchWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "fetch"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "required") {
		t.Errorf("Expected 'required' error message, got: %s", errOutput)
	}
}

func TestRunSERPWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "serp"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "required") {
		t.Errorf("Expected 'required' error message, got: %s", errOutput)
	}
}

func TestRunSuggestWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "suggest"}, stdout, stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "required") {
		t.Errorf("Expected 'required' error message, got: %s", errOutput)
	}
}

func TestRunWithDashHFlag(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "-h"}, stdout, stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message for -h, got: %s", output)
	}
}

func TestRunWithDoubleDashHelp(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"brief", "--help"}, stdout, stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message for --help, got: %s", output)
	}
}
