package brief

import (
	"flag"
	"fmt"
	"io"

	analyzepkg "github.com/code-gen-manager/brief/internal/brief/analyze"
	fetchpkg "github.com/code-gen-manager/brief/internal/brief/fetch"
	serppkg "github.com/code-gen-manager/brief/internal/brief/serp"
	suggestpkg "github.com/code-gen-manager/brief/internal/brief/suggest"
)

// Run parses CLI arguments and dispatches to specific subcommands.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)
		return 1
	}

	command := args[1]

	switch command {
	case "analyze":
		return runAnalyze(args[2:], stdout, stderr)
	case "fetch":
		return runFetch(args[2:], stdout, stderr)
	case "serp":
		return runSERP(args[2:], stdout, stderr)
	case "suggest":
		return runSuggest(args[2:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n\n", command)
		printUsage(stdout)
		return 1
	}
}

func printUsage(stdout io.Writer) {
	fmt.Fprintln(stdout, "Brief - Content Analysis Tool")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Usage: brief <command> [options]")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Commands:")
	fmt.Fprintln(stdout, "  analyze   Analyze audits and generate content gaps report")
	fmt.Fprintln(stdout, "  fetch     Fetch and parse HTML content from URL")
	fmt.Fprintln(stdout, "  serp      Fetch search engine results")
	fmt.Fprintln(stdout, "  suggest   Get search suggestions from various providers")
	fmt.Fprintln(stdout, "  help      Show this help message")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Run 'brief <command> -h' for command-specific options")
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)
	auditsFile := fs.String("audits", "", "Path do pliku all_audits.json")
	outputFile := fs.String("output", "", "Ścieżka do pliku wynikowego analysis.json")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *auditsFile == "" {
		fmt.Fprintln(stderr, "Error: audits parameter is required")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage: brief analyze -audits <file> [-output <file>]")
		return 1
	}

	opts := analyzepkg.Options{
		AuditsPath: *auditsFile,
		OutputPath: *outputFile,
	}
	if err := analyzepkg.Execute(opts, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runFetch(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	fs.SetOutput(stderr)
	urlFlag := fs.String("url", "", "URL do pobrania")
	output := fs.String("output", "", "Plik wynikowy JSON")
	method := fs.String("method", "jina", "Metoda pobierania: jina, netlify, direct")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *urlFlag == "" {
		fmt.Fprintln(stderr, "Error: url parameter is required")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage: brief fetch -url <url> [-method <jina|netlify|direct>] [-output <file>]")
		return 1
	}

	opts := fetchpkg.Options{
		URL:        *urlFlag,
		Method:     *method,
		OutputPath: *output,
	}
	if err := fetchpkg.Execute(opts, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runSERP(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	query := fs.String("query", "", "Zapytanie wyszukiwane")
	lang := fs.String("lang", "pl", "Kod języka")
	country := fs.String("country", "PL", "Kod kraju")
	limit := fs.Int("limit", 20, "Limit wyników")
	searxngURL := fs.String("searxng-url", "", "URL instancji SearXNG")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *query == "" {
		fmt.Fprintln(stderr, "Error: query parameter is required")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage: brief serp -query <query> [-lang <lang>] [-country <country>] [-limit <n>] [-searxng-url <url>]")
		return 1
	}

	opts := serppkg.Options{
		Query:      *query,
		Lang:       *lang,
		Country:    *country,
		Limit:      *limit,
		SearxngURL: *searxngURL,
	}
	if err := serppkg.Execute(opts, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runSuggest(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("suggest", flag.ContinueOnError)
	fs.SetOutput(stderr)
	provider := fs.String("provider", "google", "Dostawca: google, bing, dgd")
	query := fs.String("query", "", "Zapytanie wyszukiwane")
	lang := fs.String("lang", "pl", "Kod języka")
	country := fs.String("country", "PL", "Kod kraju")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *query == "" {
		fmt.Fprintln(stderr, "Error: query parameter is required")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage: brief suggest -query <query> [-provider <google|bing|dgd>] [-lang <lang>] [-country <country>]")
		return 1
	}

	opts := suggestpkg.Options{
		Provider: *provider,
		Query:    *query,
		Lang:     *lang,
		Country:  *country,
	}
	if err := suggestpkg.Execute(opts, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
