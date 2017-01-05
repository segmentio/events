package events

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// DefaultLogger is the default logger used by the Log function. This may be
// overwritten by the program to change the default route for log events.
var DefaultLogger = Logger{
	Handler:     Discard,
	EnablePC:    true,
	EnableDebug: true,
}

// Log emits a log event to the default logger.
func Log(format string, args ...interface{}) {
	DefaultLogger.log(1, format, args...)
}

// Debug emits a debug event to the default logger.
func Debug(format string, args ...interface{}) {
	DefaultLogger.debug(1, format, args...)
}

// A Logger is a wrapper around an event handler which exposes a Log method for
// formatting event messages before sending them to the handler.
//
// The format supported by the Log method is a superset of the fmt-style format,
// where the 'verbs' may include a column-surrounded value representing the name
// of the matching argument.
//
// The Log method also makes a special case when it gets an events.Args as last
// argument, it doesn't use it to format the message and instead simply append
// it to the event's argument list.
//
// Here's an example with the defalut logger:
//
//	events.Log("Hello %:name:!", "Luke", events.Args{
//		{"from", "Han"},
//	})
//
// Which produces an event that looks like this:
//
//	events.Event{
//		Message: "Hello Luke!",
//		Args:    events.Args{{"name", "Luke"}, {"from", "Han"}},
//		...
//	}
//
// Logger instances are safe to use concurrently from multiple goroutines.
type Logger struct {
	// Handler is the event handler receiving events from the logger.
	Handler Handler

	// Args is a list of arguments that get automatically injected into every
	// events produced by the logger.
	Args Args

	// CallDepth is used to adjust which caller the logger should report its
	// events are coming from.
	// Leaving to zero means reporting the direct caller of the logger's
	// methods.
	CallDepth int

	// EnablePC controls whether the logger should report the program counter
	// address of its caller on the events it produces.
	// This has a significant impact on the performance of the logger's Log and
	// Debug method but also provides very important insights, be mindful about
	// turning it on or off.
	EnablePC bool

	// EnableDebug controls whether calls to Debug produces events.
	EnableDebug bool
}

// Log formats an event and sends it to the logger's handler.
func (l *Logger) Log(format string, args ...interface{}) {
	l.log(1, format, args...)
}

func (l *Logger) log(depth int, format string, args ...interface{}) {
	var s = logPool.Get().(*logState)
	var a Args
	var pc [1]uintptr

	if l.EnablePC {
		runtime.Callers(l.CallDepth+depth+2, pc[:])
	}

	if n := len(args); n != 0 {
		if s, ok := args[n-1].(Args); ok {
			a, args = s, args[:n-1]
		}
	}

	s.Args = append(s.Args, l.Args...)
	s.fmt, s.Args = appendFormat(s.fmt, s.Args, format, args)
	s.Args = append(s.Args, a...)

	fmt.Fprintf(s, string(s.fmt), args...)

	if len(s.buf) != 0 {
		s.Message = *(*string)(unsafe.Pointer(&reflect.StringHeader{
			Data: uintptr(unsafe.Pointer(&s.buf[0])),
			Len:  len(s.buf),
		}))
	}

	s.Debug = l.EnableDebug
	s.PC = pc[0]
	s.Time = time.Now()
	l.Handler.HandleEvent(&s.Event)

	s.Message = ""
	s.Args = s.Args[:0]
	s.fmt = s.fmt[:0]
	s.buf = s.buf[:0]

	logPool.Put(s)
	return

}

// Debug is like Log but only produces events if the logger has debugging
// enabled.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.debug(1, format, args...)
}

func (l *Logger) debug(depth int, format string, args ...interface{}) {
	if l.EnableDebug {
		l.log(depth+1, format, args...)
	}
}

// With returns a new Logger which is a copy of l augmented with args.
func (l *Logger) With(args Args) *Logger {
	var newArgs Args
	var newLen = len(l.Args) + len(args)

	if newLen != 0 {
		newArgs = make(Args, 0, newLen)
		newArgs = append(newArgs, l.Args...)
		newArgs = append(newArgs, args...)
	}

	return &Logger{
		Args:        newArgs,
		Handler:     l.Handler,
		EnableDebug: l.EnableDebug,
	}
}

// logState is used to build events produced by Logger instances.
type logState struct {
	Event
	fmt []byte
	buf []byte
}

func (s *logState) Write(b []byte) (n int, err error) {
	s.buf = append(s.buf, b...)
	n = len(b)
	return
}

var logPool = sync.Pool{
	New: func() interface{} {
		return &logState{
			Event: Event{
				Args: make(Args, 0, 8),
			},
			fmt: make([]byte, 0, 512),
			buf: make([]byte, 0, 512),
		}
	},
}

func appendFormat(dstFmt []byte, dstArgs Args, srcFmt string, srcArgs []interface{}) ([]byte, Args) {
	for i, n := 0, len(srcFmt); i != n; {
		off := strings.IndexByte(srcFmt[i:], '%')
		if off < 0 {
			dstFmt = append(dstFmt, srcFmt[i:]...)
			break
		}
		off++
		dstFmt = append(dstFmt, srcFmt[i:i+off]...)

		if i += off; i != n && srcFmt[i] == '%' { // escaped '%'
			dstFmt = append(dstFmt, '%')
			i++
			continue
		}

		var key string
		var val interface{}
	fmtLoop:
		for i != n {
			switch c := srcFmt[i]; c {
			default:
				dstFmt = append(dstFmt, c)
				i++
				if ((c >= 'a') && (c <= 'z')) || ((c >= 'A') && (c <= 'Z')) {
					break fmtLoop
				}
				continue

			case ':': // extract the argument name from the format string
				if i++; i == n {
					dstFmt = append(dstFmt, ':')
					i = n
					break fmtLoop
				}

				j := strings.IndexByte(srcFmt[i:], ':')
				if j < 0 {
					dstFmt = append(dstFmt, srcFmt[i-1:]...)
					i = n
					break fmtLoop
				}

				key = srcFmt[i : i+j]
				i += j + 1
			}
		}

		if len(srcArgs) == 0 {
			val = missing
		} else {
			val, srcArgs = srcArgs[0], srcArgs[1:]
		}

		dstArgs = append(dstArgs, Arg{key, val})
	}

	return dstFmt, dstArgs
}

var (
	// Prevents Go from doing a memory allocation when there is a missing argument.
	missing interface{} = "MISSING"
)
