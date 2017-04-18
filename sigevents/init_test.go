package sigevents

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/segmentio/events"
)

func TestSignalHandler(t *testing.T) {
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGUSR1)
	p.Signal(syscall.SIGUSR1)

	// wait for signals to be processed, this is an asynchronous operation
	time.Sleep(10 * time.Millisecond)

	if !events.DefaultLogger.EnableDebug {
		t.Error("debug logs should be enabled after receiving SIGUSR1")
	}

	p.Signal(syscall.SIGUSR2)

	// wait for signals to be processed, this is an asynchronous operation
	time.Sleep(10 * time.Millisecond)

	if events.DefaultLogger.EnableDebug {
		t.Error("debug logs should not be enabled after receiving SIGUSR2")
	}
}
