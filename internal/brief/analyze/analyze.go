package analyze

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Options defines inputs for the analyze command.
type Options struct {
	AuditsPath string
	OutputPath string
}

// Execute loads audits, runs analysis, and writes the result JSON.
func Execute(opts Options, stdout, stderr io.Writer) error {
	if opts.AuditsPath == "" {
		return fmt.Errorf("audits path is required")
	}

	fileData, err := os.ReadFile(opts.AuditsPath)
	if err != nil {
		return fmt.Errorf("reading audits file: %w", err)
	}

	var audits []Audit
	if err := json.Unmarshal(fileData, &audits); err != nil {
		return fmt.Errorf("parsing audits JSON: %w", err)
	}

	analysis := AnalyzeAudits(audits)

	var (
		outputWriter io.Writer = stdout
		outputFile   *os.File
	)

	if opts.OutputPath != "" {
		outputFile, err = os.Create(opts.OutputPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer outputFile.Close()
		outputWriter = outputFile
	}

	encoder := json.NewEncoder(outputWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(analysis); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
