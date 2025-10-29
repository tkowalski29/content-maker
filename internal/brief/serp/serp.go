package serp

import (
	"encoding/json"
	"fmt"
	"io"
)

// Options defines the parameters used for SERP fetching.
type Options struct {
	Query      string
	Lang       string
	Country    string
	Limit      int
	SearxngURL string
}

// Execute obtains SERP results and writes JSON response.
func Execute(opts Options, stdout, stderr io.Writer) error {
	if opts.Query == "" {
		return fmt.Errorf("query is required")
	}

	results, err := FetchResults(opts.SearxngURL, opts.Query, opts.Lang, opts.Country, opts.Limit)
	if err != nil {
		return fmt.Errorf("fetching SERP: %w", err)
	}

	response := NewResponse(opts.Query, opts.Lang, opts.Country, results)

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
