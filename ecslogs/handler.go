package ecslogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/events"
)

// Handler is an event handler which formats events in a ecslogs-compatible
// format and writes them to its output.
//
// It is safe to use a handler concurrently from multiple goroutines.
type Handler struct {
	Output  io.Writer
	Program string
	Pid     int

	// synchronizes writes to the output
	mutex sync.Mutex
}

// NewHandler creates a new handler which writes to output
func NewHandler(output io.Writer) *Handler {
	return &Handler{
		Output: output,
	}
}

// HandleEvent satisfies the events.Handler interface.
func (h *Handler) HandleEvent(e *events.Event) {
	f := fmtPool.Get().(*formatter)
	f.buffer.Reset()

	f.level = "INFO"
	f.time = e.Time
	f.message = e.Message
	f.data.args = e.Args
	f.info.Source = e.Source
	f.info.Program = h.Program
	f.info.Pid = h.Pid

	if e.Debug {
		f.level = "DEBUG"
	}

	for _, a := range e.Args {
		if err, ok := a.Value.(error); ok {
			f.level = "ERROR"
			f.info.Errors = append(f.info.Errors, makeEventError(err))
		}
	}

	f.encoder.Encode(f.value)

	h.mutex.Lock()
	h.Output.Write(f.buffer.b)
	h.mutex.Unlock()

	f.info.Source = ""
	f.info.Errors = f.info.Errors[:0]
	fmtPool.Put(f)
}

type event struct {
	Level   *string    `json:"level"`
	Time    *time.Time `json:"time"`
	Info    *eventInfo `json:"info"`
	Data    *eventData `json:"data"`
	Message *string    `json:"message"`
}

type eventInfo struct {
	Program string       `json:"program,omitempty"`
	Source  string       `json:"source,omitempty"`
	Pid     int          `json:"pid,omitempty"`
	Errors  []eventError `json:"errors,omitempty"`
}

type eventError struct {
	Type  string     `json:"type,omitempty"`
	Error string     `json:"error,omitempty"`
	Errno int        `json:"errno,omitempty"`
	Stack stackTrace `json:"stack,omitempty"`
}

func makeEventError(err error) eventError {
	var cause = errors.Cause(err)
	var etype = reflect.TypeOf(cause).String()
	var error = err.Error()
	var errno = 0
	var stack stackTrace

	if se, ok := cause.(syscall.Errno); ok {
		errno = int(se)
	}

	if st, ok := err.(stackTracer); ok {
		stack = stackTrace(st.StackTrace())
	}

	return eventError{
		Type:  etype,
		Error: error,
		Errno: errno,
		Stack: stack,
	}
}

type eventData struct {
	args events.Args
}

func (data *eventData) MarshalJSON() ([]byte, error) {
	if len(data.args) == 0 {
		return []byte(`{}`), nil
	}

	b := &bytes.Buffer{}
	b.Grow(64)
	b.WriteByte('{')

	n := 0
	i := data.next(0)
	e := json.NewEncoder(b)

	for i < len(data.args) {
		if n != 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(b, `%q:`, jsonString{&data.args[i].Name})
		e.Encode(&data.args[i].Value)
		b.Truncate(b.Len() - 1) // remove trailing '\n'
		i = data.next(i + 1)
		n++
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

func (data *eventData) next(i int) int {
	for _, a := range data.args[i:] {
		if _, ok := a.Value.(error); !ok {
			break
		}
		i++
	}
	return i
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type stackTrace []errors.Frame

func (st stackTrace) MarshalJSON() ([]byte, error) {
	if len(st) == 0 {
		return []byte(`[]`), nil
	}

	b := &bytes.Buffer{}
	b.Grow(256)
	b.WriteByte('[')

	for i, frame := range st {
		if i != 0 {
			b.WriteByte(',')
		}
		pc := uintptr(frame)
		file, line := events.SourceForPC(pc)
		i++
		fmt.Fprintf(b, `"%s:%d:%s"`, file, line, funcName(file, pc))
	}

	b.WriteByte(']')
	return b.Bytes(), nil
}

func funcName(file string, pc uintptr) string {
	callers := [1]uintptr{pc}
	frames := runtime.CallersFrames(callers[:])
	f, _ := frames.Next()
	name := f.Function
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// The formatter type carries the state used during a single event formatting
// operation in HandlerEvent.
// The type is designed to allow complex memory optimizations when encoding the
// event, it actually prevents Go and especially the reflect package from doing
// any memory allocation.
type formatter struct {
	value interface{}

	level   string
	time    time.Time
	info    eventInfo
	data    eventData
	message string

	buffer  buffer
	source  buffer
	encoder json.Encoder
}

var fmtPool = sync.Pool{
	New: func() interface{} {
		f := &formatter{
			buffer: buffer{make([]byte, 0, 4096)},
			source: buffer{make([]byte, 0, 256)},
		}
		f.value = &event{
			Level:   &f.level,
			Time:    &f.time,
			Info:    &f.info,
			Data:    &f.data,
			Message: &f.message,
		}
		f.encoder = *json.NewEncoder(&f.buffer)
		f.encoder.SetEscapeHTML(false)
		return f
	},
}

// This buffer type is used as an optimization, it's faster than the standard
// bytes.Buffer because it doesn't expose such a rich API.
type buffer struct {
	b []byte
}

func (buf *buffer) Reset() {
	buf.b = buf.b[:0]
}

func (buf *buffer) Write(b []byte) (n int, err error) {
	buf.b = append(buf.b, b...)
	n = len(b)
	return
}

func (buf *buffer) WriteByte(b byte) (err error) {
	buf.b = append(buf.b, b)
	return
}

type jsonString struct{ s *string }

func (j jsonString) String() string { return *j.s }
