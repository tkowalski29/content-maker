package main

import (
	"os"

	briefcli "github.com/code-gen-manager/brief/internal/cli/brief"
)

func main() {
	os.Exit(briefcli.Run(os.Args, os.Stdout, os.Stderr))
}
