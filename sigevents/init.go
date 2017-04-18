package sigevents

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/events"
)

func init() {
	sigchan := make(chan os.Signal, 1)
	sigrecv := events.Signal(sigchan)
	signal.Notify(sigchan, syscall.SIGUSR1, syscall.SIGUSR2)

	go run(sigrecv)
}

func run(signals <-chan os.Signal) {
	// This may be seen as a race between this goroutine and the ones reading
	// the value of the EnableDebug field but in practice it doesn't matter.
	// We only care about eventual consistency here which will happen once the
	// caches of the various CPUs synchronize and will result in all debug logs
	// being enabled or disabled.
	for sig := range signals {
		switch sig {
		case syscall.SIGUSR1:
			events.DefaultLogger.EnableDebug = true
		case syscall.SIGUSR2:
			events.DefaultLogger.EnableDebug = false
		}
	}
}
