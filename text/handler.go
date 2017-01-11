package text

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/segmentio/events"
)

// DefaultTimeFormat is the default time format set on Handler.
const DefaultTimeFormat = "2006-01-02 15:04:05.000"

// Handler is an event handler which format events in a human-readable format
// and writes them to its output.
//
// It is safe to use a handler concurrently from multiple goroutines.
type Handler struct {
	Output       io.Writer      // writer receiving the formatted events
	Prefix       string         // written at the beginning of each formatted event
	TimeFormat   string         // format used for the event's time
	TimeLocation *time.Location // location to output the event time in
	EnableArgs   bool           // output detailes of each args in the events
}

// NewHandler creates a new handler which writes to output with a prefix on each
// line.
func NewHandler(prefix string, output io.Writer) *Handler {
	return &Handler{
		Output:     output,
		Prefix:     prefix,
		TimeFormat: DefaultTimeFormat,
	}
}

// HandleEvent satisfies the events.Handler interface.
func (h *Handler) HandleEvent(e *events.Event) {
	buf := bufferPool.Get().(*buffer)
	buf.b = buf.b[:0]
	buf.b = append(buf.b, h.Prefix...)

	if fmt := h.TimeFormat; len(fmt) != 0 {
		loc := h.TimeLocation
		if loc == nil {
			loc = time.Local
		}
		buf.b = e.Time.In(loc).AppendFormat(buf.b, fmt)
		buf.b = append(buf.b, " - "...)
	}

	if len(e.Source) != 0 {
		buf.b = append(buf.b, e.Source...)
		buf.b = append(buf.b, " - "...)
	}

	buf.b = append(buf.b, e.Message...)
	buf.b = append(buf.b, '\n')

	if h.EnableArgs {
		hasError := false

		for _, a := range e.Args {
			if _, ok := a.Value.(error); ok {
				hasError = true
			} else {
				buf.b = append(buf.b, '\t')
				buf.b = append(buf.b, a.Name...)
				buf.b = append(buf.b, ':', ' ')
				fmt.Fprintf(buf, "%v\n", a.Value)
			}
		}

		if hasError {
			fmt.Fprint(buf, "\terrors:\n")

			for _, a := range e.Args {
				if err, ok := a.Value.(error); ok {
					fmt.Fprintf(buf, "\t\t- %+v\n", err)
				}
			}
		}
	}

	h.Output.Write(buf.b)
	bufferPool.Put(buf)
}

// This buffer type is used as an optimization, it's faster than the standard
// bytes.Buffer because it doesn't expose such a rich API.
type buffer struct {
	b []byte
}

func (buf *buffer) Write(b []byte) (n int, err error) {
	buf.b = append(buf.b, b...)
	n = len(b)
	return
}

var bufferPool = sync.Pool{
	New: func() interface{} { return &buffer{make([]byte, 0, 4096)} },
}
