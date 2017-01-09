package log

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/segmentio/events"
	"github.com/segmentio/events/ecslogs"
	"github.com/segmentio/events/text"
)

// New creates a new Logger. The out variable sets the destination to which log
// data will be written. The prefix appears at the beginning of each generated
// log line. The flag argument defines the logging properties.
func New(out io.Writer, prefix string, flags int) *log.Logger {
	return NewLogger(prefix, flags, NewHandler(out))
}

// NewLogger creates a new logger. The prefix and flags arguments configure the
// behavior of the logger, while handler sets the destination to which log
// events are sent.
func NewLogger(prefix string, flags int, handler events.Handler) *log.Logger {
	return log.New(NewWriter(prefix, flags, handler), prefix, flags)
}

// NewHandler creates an event handler with best suits w.
//
// If w is nil the function returns nil (which a logger interprets as using the
// default handler).
func NewHandler(w io.Writer) events.Handler {
	var term bool

	if w == nil {
		return nil
	}

	if f, ok := w.(interface {
		Fd() uintptr
	}); ok {
		term = terminal.IsTerminal(int(f.Fd()))
	}

	if term {
		return text.NewHandler("", w)
	}
	return ecslogs.NewHandler(w)
}

var (
	// Cache of the output set for the default logger, we need this because the
	// standard log package doesn't expose any API to retrieve it.
	defaultWriter *Writer = NewWriter(log.Prefix(), log.Flags(), nil)
)

// =============================================================================
// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

const (
	// Bits or'ed together to control what's printed.
	// There is no control over the order they appear (the order listed
	// here) or the format they present (as described in the comments).
	// The prefix is followed by a colon only when Llongfile or Lshortfile
	// is specified.
	// For example, flags Ldate | Ltime (or LstdFlags) produce,
	//2009/01/23 01:23:23 message
	// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
	//2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
	Ldate         = log.Ldate         // the date in the local time zone: 2009/01/23
	Ltime         = log.Ltime         // the time in the local time zone: 01:23:23
	Lmicroseconds = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC          = log.LUTC          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = log.LstdFlags     // initial values for the standard logger
)

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	handler := NewHandler(w)
	defaultWriter.mutex.Lock()
	defaultWriter.handler = handler
	defaultWriter.mutex.Unlock()
}

// Flags returns the output flags for the standard logger.
func Flags() int {
	return log.Flags()
}

// SetFlags sets the output flags for the standard logger.
func SetFlags(flags int) {
	defaultWriter.mutex.Lock()
	log.SetFlags(flags)
	defaultWriter.flags = flags
	defaultWriter.mutex.Unlock()
}

// Prefix returns the output prefix for the standard logger.
func Prefix() string {
	return log.Prefix()
}

// SetPrefix sets the output prefix for the standard logger.
func SetPrefix(prefix string) {
	defaultWriter.mutex.Lock()
	log.SetPrefix(prefix)

	// Special case if the default writer is backed by a text handler so we can
	// make it output the right prefix.
	if h, ok := defaultWriter.handler.(*text.Handler); ok {
		h.Prefix = prefix
	}

	defaultWriter.prefix = prefix
	defaultWriter.mutex.Unlock()
}

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	log.Output(2, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	log.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	log.Output(2, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	log.Output(2, s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	log.Output(2, s)
	panic(s)
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func Output(calldepth int, s string) error {
	return log.Output(calldepth+1, s) // +1 for this frame.
}

// =============================================================================
