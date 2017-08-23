package log

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/segmentio/events"
	"github.com/segmentio/events/eventstest"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		prefix string
		flags  int
		format string
		args   []interface{}
		event  events.Event
	}{
		{ // empty
			prefix: "",
			flags:  0,
			format: "",
			args:   nil,
			event:  events.Event{},
		},
		{ // simple
			prefix: "==> ",
			flags:  0,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // formatted
			prefix: "==> ",
			flags:  0,
			format: "Hello %s!",
			args:   []interface{}{"Luke"},
			event: events.Event{
				Message: "Hello Luke!",
			},
		},
		{ // stdFlags
			prefix: "==> ",
			flags:  LstdFlags,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // date + microseconds + UTC
			prefix: "==> ",
			flags:  Ldate | Lmicroseconds | LUTC,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // date only
			prefix: "==> ",
			flags:  Ldate,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // time only
			prefix: "==> ",
			flags:  Ltime,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // microseconds only
			prefix: "==> ",
			flags:  Lmicroseconds,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // shortfile
			prefix: "==> ",
			flags:  Lshortfile,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // longfile
			prefix: "==> ",
			flags:  Llongfile,
			format: "Hello World!",
			args:   nil,
			event: events.Event{
				Message: "Hello World!",
			},
		},
		{ // all
			prefix: "==> ",
			flags:  Ldate | Lmicroseconds | LUTC | Llongfile,
			format: "Hello %s!",
			args:   []interface{}{"Luke"},
			event: events.Event{
				Message: "Hello Luke!",
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			h := &eventstest.Handler{}

			logger := NewLogger(test.prefix, test.flags, h)
			logger.Printf(test.format, test.args...)

			h.AssertEvents(t, test.event)
		})
	}
}

func TestNew(t *testing.T) {
	b := &bytes.Buffer{}
	l := New(b, "==> ", Ldate|Lmicroseconds|LUTC|Llongfile)
	l.Printf("Hello %s!", "Luke")

	// We can't control the time or source line this test is gonna be using so
	// we just check that the result isn't empty and output it so humans can
	// take a look at it.
	if b.Len() == 0 {
		t.Error("no bytes were produced")
	}

	t.Log(b.String())
}

func TestPrintf(t *testing.T) {
	b := &bytes.Buffer{}
	SetOutput(b)
	SetPrefix("==> ")
	SetFlags(Ldate | Lmicroseconds | LUTC | Llongfile)
	Printf("Hello %s!", "Luke")

	// We can't control the time or source line this test is gonna be using so
	// we just check that the result isn't empty and output it so humans can
	// take a look at it.
	if b.Len() == 0 {
		t.Error("no bytes were produced")
	}

	t.Log(b.String())
}

func BenchmarkLogger(b *testing.B) {
	l := New(ioutil.Discard, "==> ", Ldate|Lmicroseconds|LUTC|Llongfile)

	for i := 0; i != b.N; i++ {
		l.Printf("Hello World!")
	}
}
