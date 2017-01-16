package ecslogs

import (
	"os"
	"path/filepath"

	"github.com/segmentio/events"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	if !terminal.IsTerminal(1) {
		events.DefaultHandler = &Handler{
			Output:  os.Stdout,
			Program: filepath.Base(os.Args[0]),
			Pid:     os.Getpid(),
		}
	}
}
