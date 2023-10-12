package httpevents

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/events/v2"
	"github.com/segmentio/events/v2/eventstest"
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
	req.Header.Set("Authorization", "this will be deleted")
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
		if req.Header.Get("Authorization") != "this will be deleted" {
			t.Error("Authorization header should not change in request handler")
		}
		res.WriteHeader(http.StatusAccepted)
	}))
	h.ServeHTTP(res, req)

	eventsHandler.AssertEvents(t, events.Event{
		Message: `127.0.0.1:80->127.0.0.1:56789 - www.github.com - GET /hello?answer=42#universe - 202 Accepted - "httpevents"`,
		Args: events.Args{
			{Name: "local_address", Value: "127.0.0.1:80"},
			{Name: "remote_address", Value: "127.0.0.1:56789"},
			{Name: "host", Value: "www.github.com"},
			{Name: "method", Value: "GET"},
			{Name: "path", Value: "/hello"},
			{Name: "query", Value: "answer=42"},
			{Name: "fragment", Value: "universe"},
			{Name: "status", Value: 202},
			{Name: "request", Value: &headerList{{name: "User-Agent", value: "httpevents"}}},
			{Name: "response", Value: &headerList{}},
		},
		Debug: true,
	})
}

func filterHeaders(headers http.Header) http.Header {
	fmt.Println("before")
	fmt.Println(headers)
	headers.Del("User-Agent")
	headers.Del("local_address")
	fmt.Println("after")
	fmt.Println(headers)
	return headers
}

func mutateRequest(request2 *request) *request {
	fmt.Println("before")
	fmt.Println(request2)
	request2.path = "/mutated"
	return request2
}
func TestNewHandlerWithFormatting(t *testing.T) {
	eventsHandler := &eventstest.Handler{}

	req := httptest.NewRequest("GET", "/hello?answer=42", nil)
	req.Header.Set("User-Agent", "httpevents")
	req.Header.Set("Authorization", "this will be deleted")
	req.URL.Fragment = "universe" // for some reason NewRequest doesn't parses this
	req.Host = "www.github.com"
	req.RemoteAddr = "127.0.0.1:56789"
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, mockAddr{
		s: "127.0.0.1:80",
		n: "tcp",
	}))

	res := httptest.NewRecorder()
	log := events.NewLogger(eventsHandler)

	/*
		successHandler  = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})
		failureHandler404 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		})
	*/

	h := NewHandlerWithFormatting(mutateRequest, log, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusAccepted)
	}))
	h.ServeHTTP(res, req)

	eventsHandler.AssertEvents(t, events.Event{
		Message: `127.0.0.1:80->127.0.0.1:56789 - www.github.com - GET /mutated?answer=42#universe - 202 Accepted - "httpevents"`,
		Args: events.Args{
			{Name: "local_address", Value: "127.0.0.1:80"},
			{Name: "remote_address", Value: "127.0.0.1:56789"},
			{Name: "host", Value: "www.github.com"},
			{Name: "method", Value: "GET"},
			{Name: "path", Value: "/mutated"},
			{Name: "query", Value: "answer=42"},
			{Name: "fragment", Value: "universe"},
			{Name: "status", Value: 202},
			{Name: "request", Value: &headerList{{name: "User-Agent", value: "httpevents"}}},
			{Name: "response", Value: &headerList{}},
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
			{Name: "local_address", Value: "127.0.0.1:80"},
			{Name: "remote_address", Value: "127.0.0.1:56789"},
			{Name: "host", Value: "www.github.com"},
			{Name: "method", Value: "POST"},
			{Name: "path", Value: "/"},
			{Name: "status", Value: 500},
			{Name: "request", Value: &headerList{{name: "User-Agent", value: "httpevents"}}},
			{Name: "response", Value: &headerList{}},
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

var _ http.Hijacker = &responseWriter{}
