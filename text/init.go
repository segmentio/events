package text

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/segmentio/events"
	"golang.org/x/crypto/ssh/terminal"
)

// DefaultPrefix is used by the default handler configured when the program's
// standard output is a terminal.
//
// The value is "program-name[pid]: "
var DefaultPrefix string

func init() {
	DefaultPrefix = fmt.Sprintf("%s[%d]: ", filepath.Base(os.Args[0]), os.Getpid())

	if terminal.IsTerminal(1) {
		events.DefaultHandler = NewHandler(DefaultPrefix, os.Stdout)
	}
}
