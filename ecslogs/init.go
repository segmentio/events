package ecslogs

import (
	"os"

	"github.com/segmentio/events"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	if !terminal.IsTerminal(1) {
		events.DefaultLogger.Handler = NewHandler(os.Stdout)
	}
}
