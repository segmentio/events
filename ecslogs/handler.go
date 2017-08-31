package ecslogs

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/segmentio/events"
	"github.com/segmentio/objconv"
	"github.com/segmentio/objconv/json"
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
	f.emitter.Reset(&f.buffer)

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

	(objconv.Encoder{Emitter: &f.emitter}).Encode(f.value)
	f.buffer.WriteByte('\n')

	h.mutex.Lock()
	h.Output.Write(f.buffer.b)
	h.mutex.Unlock()

	f.info.Source = ""
	f.info.Errors = f.info.Errors[:0]
	fmtPool.Put(f)
}

type event struct {
	Level   *string    `objconv:"level"`
	Time    *time.Time `objconv:"time"`
	Info    *eventInfo `objconv:"info"`
	Data    *eventData `objconv:"data"`
	Message *string    `objconv:"message"`
}

type eventInfo struct {
	Program string       `objconv:"program,omitempty"`
	Source  string       `objconv:"source,omitempty"`
	Pid     int          `objconv:"pid,omitempty"`
	Errors  []eventError `objconv:"errors,omitempty"`
}

type eventError struct {
	Type  string     `objconv:"type,omitempty"`
	Error string     `objconv:"error,omitempty"`
	Errno int        `objconv:"errno,omitempty"`
	Stack stackTrace `objconv:"stack,omitempty"`
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

func (data *eventData) EncodeValue(e objconv.Encoder) error {
	n := len(data.args)
	i := data.next(0)
	return e.EncodeMap(-1, func(k objconv.Encoder, v objconv.Encoder) (err error) {
		if i != n {
			if err = k.Encode(&data.args[i].Name); err != nil {
				return
			}
			if err = v.Encode(&data.args[i].Value); err != nil {
				return
			}
			i = data.next(i + 1)
		}
		if i == n {
			err = objconv.End
		}
		return
	})
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

func (st stackTrace) EncodeValue(e objconv.Encoder) error {
	f := fmtPool.Get().(*formatter)
	i := 0

	err := e.EncodeArray(len(st), func(e objconv.Encoder) error {
		f.buffer.Reset()
		pc := uintptr(st[i])
		file, line := events.SourceForPC(pc)
		i++
		fmt.Fprintf(&f.buffer, "%s:%d:%s", file, line, funcName(file, pc))
		return e.Encode(stringNoCopy(f.buffer.b))
	})

	fmtPool.Put(f)
	return err
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
	emitter json.Emitter
}

var fmtPool = sync.Pool{
	New: func() interface{} {
		f := &formatter{
			buffer:  buffer{make([]byte, 0, 4096)},
			source:  buffer{make([]byte, 0, 256)},
			emitter: *json.NewEmitter(nil),
		}
		f.value = &event{
			Level:   &f.level,
			Time:    &f.time,
			Info:    &f.info,
			Data:    &f.data,
			Message: &f.message,
		}
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

func stringNoCopy(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&b[0])),
		Len:  len(b),
	}))
}
