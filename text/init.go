// +build !testing

package text

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/segmentio/events"
)

func init() {

	DefaultPrefix = fmt.Sprintf("%s[%d]: ", filepath.Base(os.Args[0]), os.Getpid())

	if events.IsTerminal(1) {
		events.DefaultHandler = NewHandler(DefaultPrefix, os.Stdout)
	}
}
