package gradio

import (
	"flag"
	"fmt"
	"io"

	gradiopkg "github.com/code-gen-manager/brief/internal/gradio"
)

// Run obsługuje CLI dla generatora obrazów Gradio.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{"gradio"}
	}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	inputFile := fs.String("input", "", "Plik JSON z żądaniami obrazów")
	gradioURL := fs.String("gradio_url", "", "URL serwera Gradio")
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	if *inputFile == "" || *gradioURL == "" {
		fmt.Fprintln(stdout, "Usage: gradio -input=<json_file> -gradio_url=<url>")
		return 1
	}

	runner := gradiopkg.NewRunner(stdout, stderr)
	results, err := runner.RunFile(*inputFile, *gradioURL)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	exitCode := 0
	for _, res := range results {
		if res.Err != nil {
			exitCode = 1
			break
		}
	}

	return exitCode
}
