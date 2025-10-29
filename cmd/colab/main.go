package main

import (
	"os"

	colabcli "github.com/code-gen-manager/brief/internal/cli/colab"
)

func main() {
	os.Exit(colabcli.Run(os.Args, os.Stdout, os.Stderr))
}
