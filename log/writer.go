package log

import (
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/segmentio/events/v2"
)

// Writer is an implementation of an io.Writer which is designed to be set as
// output on a logger from the standard log package.
//
// The Writer parses the lines it receives when its Write method is called and
// constructs events that are then passed to its handler.
//
// It is safe to use an Writer from multiple goroutines. However a single writer
// should be associated with a single logger, the behavior is undefined.
type Writer struct {
	handler events.Handler // the handler to send the events to
	prefix  string         // the prefix set on the logger that uses this output
	flags   int            // the flags set on the logger that uses this output
	mutex   sync.Mutex     // the mutex used to control access to the writer
}

// NewWriter constructs a new Writer value which is intended to be set on a
// logger configured with prefix and flags, the output forwards events to the
// given handler.
func NewWriter(prefix string, flags int, handler events.Handler) *Writer {
	return &Writer{
		handler: handler,
		prefix:  prefix,
		flags:   flags,
	}
}

// Write satisfies the io.Writer interface.
func (w *Writer) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	var h = w.handler
	var e = eventPool.Get().(*events.Event)
	var s = *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&b[0])),
		Len:  len(b),
	}))
	var t time.Time
	var src string

	if h == nil {
		h = events.DefaultHandler
	}

	w.mutex.Lock()
	flags, prefix := w.flags, w.prefix

	if strings.HasPrefix(s, prefix) {
		s = s[len(prefix):]
	}

	if format := timeFormat(flags); len(format) != 0 {
		var tz *time.Location
		var ts string

		if n := len(format); len(s) >= n {
			ts, s = s[:n], s[n:]
		} else {
			ts, s = s, ""
		}

		if (flags & LUTC) != 0 {
			tz = time.UTC
		} else {
			tz = time.Local
		}

		t, _ = time.ParseInLocation(format, ts, tz)
	}

	if (flags & (Llongfile | Lshortfile)) != 0 {
		var base string
		var file string
		var line string

		s = skip(s, ' ')
		base = s
		file, s = parse(s, ':') // parse the file name or path
		line, s = parse(s, ':') // parse the line number

		src = base[:len(file)+len(line)+1]
	}

	s = skip(s, ' ')

	if n := strings.IndexByte(s, '\n'); n >= 0 {
		s = s[:n]
	}

	if t == (time.Time{}) {
		t = time.Now()
	}

	e.Message = s
	e.Source = src
	e.Time = t

	h.HandleEvent(e)
	w.mutex.Unlock()

	e.Message = ""
	e.Source = ""
	e.Time = time.Time{}
	eventPool.Put(e)

	return len(b), nil
}

func timeFormat(flags int) string {
	switch flags & (Ldate | Ltime | Lmicroseconds) {
	case Ldate | Ltime | Lmicroseconds, Ldate | Lmicroseconds:
		return "2006/01/02 15:04:05.999999"
	case Ldate | Ltime:
		return "2006/01/02 15:04:05"
	case Ldate:
		return "2006/01/02"
	case Ltime | Lmicroseconds, Lmicroseconds:
		return "15:04:05.999999"
	case Ltime:
		return "15:04:05"
	default:
		return ""
	}
}

func skip(s string, b byte) string {
	if len(s) != 0 && s[0] == b {
		s = s[1:]
	}
	return s
}

func parse(s string, b byte) (left string, right string) {
	if index := strings.IndexByte(s, b); index >= 0 {
		left, right = s[:index], s[index+1:]
	} else {
		left = s
	}
	return
}

var eventPool = sync.Pool{
	New: func() interface{} { return &events.Event{} },
}
