package events

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// Signal triggers events on handler for every signal that arrives on sigchan.
// The function returns a channel on which signals are forwarded, the program
// should use this channel instead of sigchan to receive signals.
func Signal(sigchan <-chan os.Signal, handler Handler) <-chan os.Signal {
	output := make(chan os.Signal)

	// We capture the stack frame here instead of in the goroutine because it
	// gives a more meaningful value (the caller of Signal, which usually is
	// the application itself).
	var pc [1]uintptr
	runtime.Callers(2, pc[:])
	file, line := SourceForPC(pc[0])

	go func() {
		for sig := range sigchan {
			h := handler

			if h == nil {
				h = DefaultHandler
			}

			h.HandleEvent(&Event{
				Message: sig.String(),
				Source:  fmt.Sprintf("%s:%d", file, line),
				Time:    time.Now(),
				Args:    Args{{"signal", sig}},
			})

			// Limits to 1s the attempt to publish to the output channel, this
			// is a safeguard for programs that don't consume from the output
			// channel (event though they should).
			select {
			case output <- sig:
			case <-time.After(time.Second):
			}
		}
	}()

	return output
}
