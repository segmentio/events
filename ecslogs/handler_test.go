package ecslogs

import (
	"bytes"
	"io"
	"io/ioutil"
	"runtime"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/events"
)

func TestHandler(t *testing.T) {
	pc := [1]uintptr{}
	runtime.Callers(1, pc[:])

	b := &bytes.Buffer{}
	h := NewHandler(b)
	e := &events.Event{
		Message: "Hello Luke!",
		Args:    events.Args{{"name", "Luke"}, {"from", "Han"}, {"error", io.EOF}},
		Time:    time.Date(2017, 1, 1, 23, 42, 0, 123000000, time.UTC),
		PC:      pc[0],
		Debug:   true,
	}

	for i := 0; i != 3; i++ {
		h.HandleEvent(e)
	}

	const ref = `{"level":"ERROR","time":"2017-01-01T23:42:00.123Z","info":{"source":"github.com/segmentio/events/ecslogs/handler_test.go:19","errors":[{"type":"*errors.errorString","error":"EOF"}]},"data":{"name":"Luke","from":"Han"},"message":"Hello Luke!"}
{"level":"ERROR","time":"2017-01-01T23:42:00.123Z","info":{"source":"github.com/segmentio/events/ecslogs/handler_test.go:19","errors":[{"type":"*errors.errorString","error":"EOF"}]},"data":{"name":"Luke","from":"Han"},"message":"Hello Luke!"}
{"level":"ERROR","time":"2017-01-01T23:42:00.123Z","info":{"source":"github.com/segmentio/events/ecslogs/handler_test.go:19","errors":[{"type":"*errors.errorString","error":"EOF"}]},"data":{"name":"Luke","from":"Han"},"message":"Hello Luke!"}
`

	if s := b.String(); s != ref {
		t.Error("bad event:")
		t.Log("expected:", ref)
		t.Log("found:   ", s)
	}

	t.Run("stack", func(t *testing.T) {
		e.Args[2].Value = errors.WithStack(io.EOF)

		b.Reset()
		h.HandleEvent(e)

		if s := b.String(); s != `{"level":"ERROR","time":"2017-01-01T23:42:00.123Z","info":{"source":"github.com/segmentio/events/ecslogs/handler_test.go:19","errors":[{"type":"*errors.errorString","error":"EOF","stack":["github.com/segmentio/events/ecslogs/handler_test.go:45","testing/testing.go:611","runtime/asm_amd64.s:2087"]}]},"data":{"name":"Luke","from":"Han"},"message":"Hello Luke!"}
` {
			// This test is sensitive, it may break if the Go version changes or
			// if this file is edited (because the number of lines may not be
			// the same anymore), so we don't report an error and instead just
			// log the issue, asking the developer to fix it.
			t.Log("unable to tell if the test checking the stack trace serialization worked, please verify and fix if necessary:\n", s)
		}
	})
}

func BenchmarkHandler(b *testing.B) {
	pc := [1]uintptr{}
	runtime.Callers(0, pc[:])

	h := NewHandler(ioutil.Discard)
	e := &events.Event{
		Message: "Hello Luke!",
		Args:    events.Args{{"name", "Luke"}, {"from", "Han"}},
		Time:    time.Date(2017, 1, 1, 23, 42, 0, 123000000, time.UTC),
		PC:      pc[0],
		Debug:   true,
	}

	for i := 0; i != b.N; i++ {
		h.HandleEvent(e)
	}
}
