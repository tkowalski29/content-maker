package suggest

import (
	"encoding/json"
	"fmt"
	"io"
)

// Options holds parameters for suggestions lookup.
type Options struct {
	Provider string
	Query    string
	Lang     string
	Country  string
}

// Execute retrieves search suggestions and writes JSON response.
func Execute(opts Options, stdout, stderr io.Writer) error {
	if opts.Query == "" {
		return fmt.Errorf("query is required")
	}

	suggestions, err := FetchSuggestions(opts.Provider, opts.Query, opts.Lang, opts.Country)
	if err != nil {
		return fmt.Errorf("fetching suggestions: %w", err)
	}

	response := NewResponse(opts.Provider, opts.Query, suggestions)

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
