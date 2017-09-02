package httpevents

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/events"
	"github.com/segmentio/events/eventstest"
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
	eventsHandler := &eventstest.Handler{}

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
	log := events.NewLogger(eventsHandler)

	h := NewHandlerWith(log, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusAccepted)
	}))
	h.ServeHTTP(res, req)

	eventsHandler.AssertEvents(t, events.Event{
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
			{"request-headers", &headerList{{"User-Agent", "httpevents"}}},
			{"response-headers", &headerList{}},
		},
		Debug: true,
	})
}

func TestHandlerPanic(t *testing.T) {
	eventsHandler := &eventstest.Handler{}

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("User-Agent", "httpevents")
	req.Host = "www.github.com"
	req.RemoteAddr = "127.0.0.1:56789"
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, mockAddr{
		s: "127.0.0.1:80",
		n: "tcp",
	}))

	res := httptest.NewRecorder()
	log := events.NewLogger(eventsHandler)

	h := NewHandlerWith(log, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		panic("bye bye!")
	}))

	func() {
		defer func() { recover() }()
		h.ServeHTTP(res, req)
	}()

	eventsHandler.AssertEvents(t, events.Event{
		Message: `127.0.0.1:80->127.0.0.1:56789 - www.github.com - POST / - 500 Internal Server Error - "httpevents"`,
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"host", "www.github.com"},
			{"method", "POST"},
			{"path", "/"},
			{"status", 500},
			{"request-headers", &headerList{{"User-Agent", "httpevents"}}},
			{"response-headers", &headerList{}},
		},
	})
}

func BenchmarkHandler(b *testing.B) {
	l := &events.Logger{}
	h := NewHandlerWith(l, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))

	w := &mockResponseWriter{}
	r := httptest.NewRequest("GET", "http://localhost:4242/", nil)

	for i := 0; i != b.N; i++ {
		h.ServeHTTP(w, r)
	}
}

type mockResponseWriter struct{}

func (w *mockResponseWriter) WriteHeader(status int)      {}
func (w *mockResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *mockResponseWriter) Header() http.Header         { return http.Header{} }

type mockAddr struct {
	s string
	n string
}

func (a mockAddr) String() string  { return a.s }
func (a mockAddr) Network() string { return a.n }

var (
	_ http.Hijacker = &responseWriter{}
)
