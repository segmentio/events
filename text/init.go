//go:build !testing
// +build !testing

package text

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/segmentio/events/v2"
	"golang.org/x/term"
)

func init() {
	DefaultPrefix = fmt.Sprintf("%s[%d]: ", filepath.Base(os.Args[0]), os.Getpid())

	if term.IsTerminal(1) {
		events.DefaultHandler = NewHandler(DefaultPrefix, os.Stdout)
	}
}
