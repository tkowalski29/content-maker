package colab

import (
	"io"
	"log"

	colabrunner "github.com/code-gen-manager/brief/internal/colab"
)

func Run(args []string, _ io.Writer, stderr io.Writer) int {
	if stderr != nil {
		log.SetOutput(stderr)
	}
	return colabrunner.Run(args)
}
