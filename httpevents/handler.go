package httpevents

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/segmentio/events/v2"
)

// NewHandler wraps the HTTP handler and returns a new handler which logs all 4xx
// and 5xx requests with the default logger. All other requests will only be logged
// in debug mode.
func NewHandler(handler http.Handler) http.Handler {
	return NewHandlerWith(events.DefaultLogger, handler)
}

// NewHandlerWith wraps the HTTP handler and returns a new handler which logs all
// requests with logger.
//
// Panics from handler are intercepted and trigger a 500 response if no response
// header was sent yet. The panic is not silenced tho and is propagated to the
// parent handler.
func NewHandlerWith(logger *events.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var laddr string

		if value, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			laddr = value.String()
		}

		w := responseWriterPool.Get().(*responseWriter)
		// We capture all the values we need from req in case the object
		// gets modified by the handler.
		w.ResponseWriter = res
		w.logger = logger
		w.request.reset(req, laddr)

		// If the handler panics we want to make sure we report the issue in the
		// access log, while also ensuring that a response is going to be sent
		// down to the client.
		// We don't silence the panic here tho and instead we forward it back to
		// the parent handler which may need to be aware that a panic occurred.
		defer func() {
			err := recover()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.ResponseWriter = nil
			w.logger = nil
			w.wroteHeader = false
			w.request.release()
			responseWriterPool.Put(w)

			if err != nil {
				panic(err)
			}
		}()

		// The request is forwarded to the handler, if it never calls the
		// writer's WriteHeader method we force the call with "200 OK" status
		// to match the default behavior of the net/http package (and also make
		// sure an access log will be written).
		handler.ServeHTTP(w, req)
		w.WriteHeader(http.StatusOK)
	})
}

type LoggerFunc func(headers http.Header) http.Header

func copyHeaders(headers http.Header) http.Header {
	fmt.Println("Copying headers")
	fmt.Println("Existing Header")
	for k, v := range headers {
		fmt.Println(k)
		fmt.Println(v)
	}
	headersCopy := make(http.Header)

	for k, v := range headers {
		headersCopy[k] = v
	}

	fmt.Println("New Headers")
	for k, v := range headersCopy {
		fmt.Println(k)
		fmt.Println(v)
	}

	return headersCopy
}
func NewHandlerWithFormatting(formatter RequestTransformFunc, logger *events.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var laddr string

		if value, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			laddr = value.String()
		}

		w := responseWriterPool.Get().(*responseWriter)
		// We capture all the values we need from req in case the object
		// gets modified by the handler.
		w.ResponseWriter = res
		w.logger = logger
		w.MutateRequest = formatter
		w.request.reset(req, laddr)

		// If the handler panics we want to make sure we report the issue in the
		// access log, while also ensuring that a response is going to be sent
		// down to the client.
		// We don't silence the panic here tho and instead we forward it back to
		// the parent handler which may need to be aware that a panic occurred.
		defer func() {
			err := recover()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.ResponseWriter = nil
			w.logger = nil
			w.wroteHeader = false
			w.request.release()
			responseWriterPool.Put(w)

			if err != nil {
				panic(err)
			}
		}()

		// The request is forwarded to the handler, if it never calls the
		// writer's WriteHeader method we force the call with "200 OK" status
		// to match the default behavior of the net/http package (and also make
		// sure an access log will be written).
		handler.ServeHTTP(w, req)
		w.WriteHeader(http.StatusOK)
	})

}

type RequestTransformFunc func(request2 *request) *request
type responseWriter struct {
	http.ResponseWriter
	logger *events.Logger
	request
	wroteHeader     bool
	SanitizeHeaders LoggerFunc
	MutateRequest   RequestTransformFunc
}

func (w *responseWriter) WriteHeader(status int) {
	fmt.Println("Headers before logging")
	for k, v := range w.request.reqHeaders {
		fmt.Println(k)
		fmt.Println(v)
	}
	w.log(1, status)

	if !w.wroteHeader {
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) Hijack() (conn net.Conn, rw *bufio.ReadWriter, err error) {
	if conn, rw, err = w.ResponseWriter.(http.Hijacker).Hijack(); err == nil {
		w.log(1, http.StatusSwitchingProtocols)
	}
	return
}

func (w *responseWriter) log(depth int, status int) {
	if logger := w.logger; logger != nil {
		w.logger = nil
		w.request.status = status
		w.request.statusText = http.StatusText(status)

		if w.SanitizeHeaders != nil {
			fmt.Println("Sanitizing headers")
			for k, v := range w.ResponseWriter.Header() {
				fmt.Println(k)
				fmt.Println(v)
			}
			headers := w.SanitizeHeaders(copyHeaders(w.ResponseWriter.Header()))
			fmt.Println("Headers")
			println(w.ResponseWriter.Header())

			println(w.ResponseWriter.Header().Get("User-Agent"))
			fmt.Println("Changed Headers")
			println(headers)
			println(headers.Get("User-Agent"))
			w.request.log(logger, headers, depth+1, w.MutateRequest)
		} else {
			fmt.Println("MISSING FUNC")
			w.request.log(logger, w.ResponseWriter.Header(), depth+1, w.MutateRequest)
		}
	}
}

var responseWriterPool = sync.Pool{
	New: func() interface{} { return &responseWriter{} },
}
