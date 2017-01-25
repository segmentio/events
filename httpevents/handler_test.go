package httpevents

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events"
)

type writer struct {
	http.ResponseWriter
	writeHeader int
}

func (w *writer) WriteHeader(status int) {
	w.writeHeader++
	w.ResponseWriter.WriteHeader(status)
}

func TestHandlerWriteOnce(t *testing.T) {
	w := &writer{
		ResponseWriter: httptest.NewRecorder(),
	}
	r := &responseWriter{
		ResponseWriter: w,
	}

	r.WriteHeader(http.StatusOK)
	r.WriteHeader(http.StatusOK)

	if w.writeHeader != 1 {
		t.Error("invalid number of WriteHeader calls to the base ResponseWriter:", w.writeHeader)
	}
}

func TestHandler(t *testing.T) {
	evList := []*events.Event{}

	req := httptest.NewRequest("GET", "/hello?answer=42", nil)
	req.Header.Set("User-Agent", "httpevents")
	req.URL.Fragment = "universe" // for some reason NewRequest doesn't parses this
	req.Host = "www.github.com"
	req.RemoteAddr = "127.0.0.1:56789"
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, mockAddr{
		s: "127.0.0.1:80",
		n: "tcp",
	}))

	res := httptest.NewRecorder()
	log := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	h := NewHandler(log, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusAccepted)
	}))
	h.ServeHTTP(res, req)

	if len(evList) != 1 {
		t.Error("bad event count:", len(evList))
		return
	}

	e := *evList[0]
	e.Source = ""
	e.Time = time.Time{}

	if !reflect.DeepEqual(e, events.Event{
		Message: `127.0.0.1:80->127.0.0.1:56789 - www.github.com - GET /hello?answer=42#universe - 202 Accepted - "httpevents"`,
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"host", "www.github.com"},
			{"method", "GET"},
			{"path", "/hello"},
			{"query", "answer=42"},
			{"fragment", "universe"},
			{"status", 202},
			{"agent", "httpevents"},
		},
	}) {
		t.Errorf("%#v", e)
	}
}

func TestHandlerPanic(t *testing.T) {
	evList := []*events.Event{}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("User-Agent", "httpevents")
	req.Host = "www.github.com"
	req.RemoteAddr = "127.0.0.1:56789"
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, mockAddr{
		s: "127.0.0.1:80",
		n: "tcp",
	}))

	res := httptest.NewRecorder()
	log := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	h := NewHandler(log, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		panic("bye bye!")
	}))

	func() {
		defer func() { recover() }()
		h.ServeHTTP(res, req)
	}()

	if len(evList) != 1 {
		t.Error("bad event count:", len(evList))
		return
	}

	e := *evList[0]
	e.Source = ""
	e.Time = time.Time{}

	if !reflect.DeepEqual(e, events.Event{
		Message: `127.0.0.1:80->127.0.0.1:56789 - www.github.com - POST / - 500 Internal Server Error - "httpevents"`,
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"host", "www.github.com"},
			{"method", "POST"},
			{"path", "/"},
			{"status", 500},
			{"agent", "httpevents"},
		},
	}) {
		t.Errorf("%#v", e)
	}
}

type mockAddr struct {
	s string
	n string
}

func (a mockAddr) String() string  { return a.s }
func (a mockAddr) Network() string { return a.n }

var (
	_ http.Hijacker = &responseWriter{}
)
