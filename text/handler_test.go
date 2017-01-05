package text

import (
	"bytes"
	"io"
	"io/ioutil"
	"runtime"
	"testing"
	"time"

	"github.com/segmentio/events"
)

func TestHandler(t *testing.T) {
	pc := [2]uintptr{}
	runtime.Callers(1, pc[:])

	b := &bytes.Buffer{}
	h := NewHandler("==>", b)

	h.HandleEvent(&events.Event{
		Message: "Hello Luke!",
		Args:    events.Args{{"name", "Luke"}, {"from", "Han"}, {"error", io.EOF}},
		Time:    time.Date(2017, 1, 1, 23, 42, 0, 123000000, time.Local),
		PC:      pc[0],
		Debug:   true,
	})

	if s := b.String(); s != `==> 2017-01-01 23:42:00.123 - github.com/segmentio/events/text/handler_test.go:18 - Hello Luke!
	name: Luke
	from: Han
	errors:
		- EOF
` {
		t.Error(s)
	}
}

func BenchmarkHandler(b *testing.B) {
	pc := [2]uintptr{}
	runtime.Callers(1, pc[:])

	h := NewHandler("", ioutil.Discard)
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
