// +build !testing

package ecslogs

import (
	"os"
	"path/filepath"

	"github.com/segmentio/events"
)

func init() {

	if !events.IsTerminal(1) {
		events.DefaultHandler = &Handler{
			Output:  os.Stdout,
			Program: filepath.Base(os.Args[0]),
			Pid:     os.Getpid(),
		}
	}
}
