//go:build !testing
// +build !testing

package ecslogs

import (
	"os"
	"path/filepath"

	"github.com/segmentio/events/v2"
	"golang.org/x/term"
)

func init() {
	if !term.IsTerminal(1) {
		events.DefaultHandler = &Handler{
			Output:  os.Stdout,
			Program: filepath.Base(os.Args[0]),
			Pid:     os.Getpid(),
		}
	}
}
