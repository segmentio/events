package text

import (
	"os"

	"github.com/segmentio/events"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	if terminal.IsTerminal(1) {
		events.DefaultHandler = NewHandler("", os.Stdout)
	}
}
