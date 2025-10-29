package main

import (
	"os"

	extractcli "github.com/code-gen-manager/brief/internal/cli/extract"
)

func main() {
	os.Exit(extractcli.Run(os.Args, os.Stdout, os.Stderr))
}
