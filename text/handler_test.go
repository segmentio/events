package text

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/segmentio/events"
)

func TestHandler(t *testing.T) {
	b := &bytes.Buffer{}
	h := NewHandler("==> ", b)

	h.HandleEvent(&events.Event{
		Message: "Hello Luke!",
		Source:  "github.com/segmentio/events/text/handler_test.go:18",
		Args:    events.Args{{"name", "Luke"}, {"from", "Han"}, {"error", io.EOF}},
		Time:    time.Date(2017, 1, 1, 23, 42, 0, 123000000, time.Local),
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
	h := NewHandler("", ioutil.Discard)
	e := &events.Event{
		Message: "Hello Luke!",
		Source:  "github.com/segmentio/events/text/handler_test.go:18",
		Args:    events.Args{{"name", "Luke"}, {"from", "Han"}},
		Time:    time.Date(2017, 1, 1, 23, 42, 0, 123000000, time.UTC),
		Debug:   true,
	}

	for i := 0; i != b.N; i++ {
		h.HandleEvent(e)
	}
}
