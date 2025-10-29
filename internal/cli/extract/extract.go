package extract

import (
	"flag"
	"fmt"
	"io"

	"github.com/code-gen-manager/brief/internal/extractor"
)

// Run handles top-level dispatch for extract commands.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)
		return 1
	}

	command := args[1]

	switch command {
	case "images":
		return runImages(args[2:], stdout, stderr)
	case "frontmatter":
		return runFrontmatter(args[2:], stdout, stderr)
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
	fmt.Fprintln(stdout, "Extract - Content Extraction Tool")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Usage: extract <command> [options]")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Commands:")
	fmt.Fprintln(stdout, "  images       Extract image placeholders from markdown")
	fmt.Fprintln(stdout, "  frontmatter  Extract YAML front-matter to JSON")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Examples:")
	fmt.Fprintln(stdout, "  extract images -input=article.md -output=images.json")
	fmt.Fprintln(stdout, "  extract frontmatter -input=article.md -output=cms_data.json")
}

func runImages(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("images", flag.ContinueOnError)
	fs.SetOutput(stderr)
	inputFile := fs.String("input", "", "Input markdown file")
	outputFile := fs.String("output", "", "Output JSON file")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *inputFile == "" || *outputFile == "" {
		fmt.Fprintln(stderr, "Usage: extract images -input=<file.md> -output=<file.json>")
		return 1
	}

	count, err := extractor.ExtractImages(extractor.ImageOptions{
		InputPath:  *inputFile,
		OutputPath: *outputFile,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stderr, "✅ Extracted %d images to %s\n", count, *outputFile)
	return 0
}

func runFrontmatter(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("frontmatter", flag.ContinueOnError)
	fs.SetOutput(stderr)
	inputFile := fs.String("input", "", "Input markdown file")
	outputFile := fs.String("output", "", "Output JSON file")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *inputFile == "" || *outputFile == "" {
		fmt.Fprintln(stderr, "Usage: extract frontmatter -input=<file.md> -output=<file.json>")
		return 1
	}

	if err := extractor.ExtractFrontmatter(extractor.FrontmatterOptions{
		InputPath:  *inputFile,
		OutputPath: *outputFile,
	}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stderr, "✅ Extracted CMS data to %s\n", *outputFile)
	return 0
}
