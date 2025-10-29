package main

import (
	"os"

	gradiocli "github.com/code-gen-manager/brief/internal/cli/gradio"
)

func main() {
	os.Exit(gradiocli.Run(os.Args, os.Stdout, os.Stderr))
}
