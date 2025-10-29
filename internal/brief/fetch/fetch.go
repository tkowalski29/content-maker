package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Options encapsulates parameters for fetching HTML.
type Options struct {
	URL        string
	Method     string
	OutputPath string
}

// Execute runs the fetch process and writes JSON response.
func Execute(opts Options, stdout, stderr io.Writer) error {
	if opts.URL == "" {
		return fmt.Errorf("url is required")
	}

	result, err := FetchHTML(opts.URL, opts.Method)
	if err != nil {
		return fmt.Errorf("fetching URL: %w", err)
	}

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
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
