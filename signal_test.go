package events

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	evlist := []*Event{}
	sigchan := make(chan os.Signal, 1)
	handler := HandlerFunc(func(e *Event) { evlist = append(evlist, e.Clone()) })

	sigchan <- os.Interrupt
	output := SignalWith(handler, sigchan)

	select {
	case sig := <-output:
		if sig != os.Interrupt {
			t.Error("bad signal received:", sig)
		}
	case <-time.After(time.Second):
		t.Error("no signal received after 1s")
	}

	if len(evlist) != 1 {
		t.Error("bad result count:", len(evlist))
		return
	}

	if evlist[0].Source == "" {
		t.Error("missing source in generated event")
	}

	if evlist[0].Time == (time.Time{}) {
		t.Error("missing time in generated event")
	}

	// Unpredictable values.
	evlist[0].Source = ""
	evlist[0].Time = time.Time{}

	if !reflect.DeepEqual(*evlist[0], Event{
		Message: "interrupt",
		Args:    Args{{"signal", os.Interrupt}},
	}) {
		t.Errorf("bad event: %#v", *evlist[0])
	}
}

func TestWithSignals(t *testing.T) {
	t.Run("cancelling returns context.Canceled", func(t *testing.T) {
		ctx, cancel := WithSignals(context.Background(), os.Interrupt)

		// Ensure we can call cancel multiple times
		cancel()
		cancel()

		select {
		case <-ctx.Done():
		default:
			t.Error("the context should have been canceled after the cancellation function was called")
			return
		}

		if err := ctx.Err(); err != context.Canceled {
			t.Error("bad error returned after the context was canceled:", err)
		}
	})

	t.Run("receive os.Interrupt", func(t *testing.T) {
		ctx, cancel := WithSignals(context.Background(), os.Interrupt)
		defer cancel()

		p, _ := os.FindProcess(os.Getpid())
		p.Signal(os.Interrupt)

		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Error("no signals received within 1 second")
			return
		}

		err := ctx.Err()

		switch e := err.(type) {
		case *SignalError:
			if e.Signal != os.Interrupt {
				t.Error("bad signal returned by the context:", e)
			}
		default:
			t.Error("bad error returned by the context:", e)
		}
	})

	t.Run("report cancellation of the parent context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sig, stop := WithSignals(ctx, os.Interrupt)
		defer stop()

		cancel()

		select {
		case <-sig.Done():
		case <-time.After(1 * time.Second):
			t.Error("no cancellation received within 1 second")
			return
		}

		if err := sig.Err(); err != context.Canceled {
			t.Error("the parent error wasn't reported:", err)
		}
	})
}
