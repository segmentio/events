package events

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Signal triggers events on the default handler for every signal that arrives
// on sigchan.
// The function returns a channel on which signals are forwarded, the program
// should use this channel instead of sigchan to receive signals.
func Signal(sigchan <-chan os.Signal) <-chan os.Signal {
	return SignalWith(DefaultHandler, sigchan)
}

// SignalWith triggers events on handler for every signal that arrives on
// sigchan.
// The function returns a channel on which signals are forwarded, the program
// should use this channel instead of sigchan to receive signals.
func SignalWith(handler Handler, sigchan <-chan os.Signal) <-chan os.Signal {
	output := make(chan os.Signal)

	// We capture the stack frame here instead of in the goroutine because it
	// gives a more meaningful value (the caller of Signal, which usually is
	// the application itself).
	var pc [1]uintptr
	runtime.Callers(2, pc[:])
	file, line := SourceForPC(pc[0])

	go func() {
		defer close(output)

		for sig := range sigchan {
			handler.HandleEvent(&Event{
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

// WithSignals returns a copy of the given context which may be canceled if any
// of the given signals is received by the program.
func WithSignals(ctx context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	done := make(chan struct{})
	sigchan := make(chan os.Signal)
	sigrecv := Signal(sigchan)
	signal.Notify(sigchan, signals...)

	sig := &signalCtx{
		parent: ctx,
		done:   done,
	}

	go func() {
		select {
		case s := <-sigrecv:
			sig.cancel(&SignalError{Signal: s})
		case <-ctx.Done():
			sig.cancel(ctx.Err())
		}
		signal.Stop(sigchan)
	}()

	return sig, func() { signal.Stop(sigchan); sig.cancel(context.Canceled); close(sigchan) }
}

// SignalError is a wrapper for the os.Signal type which also implements the
// error interface so it can be reported by the Err method of a context.
type SignalError struct {
	os.Signal
}

// Error satisfies the error interface.
func (s *SignalError) Error() string {
	return s.String()
}

type signalCtx struct {
	parent context.Context
	once   sync.Once
	err    atomic.Value
	done   chan struct{}
}

func (s *signalCtx) Deadline() (time.Time, bool) {
	return s.parent.Deadline()
}

func (s *signalCtx) Done() <-chan struct{} {
	return s.done
}

func (s *signalCtx) Err() error {
	err, _ := s.err.Load().(error)
	return err
}

func (c *signalCtx) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

func (s *signalCtx) cancel(err error) {
	s.once.Do(func() {
		s.err.Store(err)
		close(s.done)
	})
}
