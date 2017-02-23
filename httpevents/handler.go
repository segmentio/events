package httpevents

import (
	"bufio"
	"net"
	"net/http"

	"github.com/segmentio/events"
)

// NewHandler wraps the HTTP handler and returns a new handler which logs all
// requests with the default logger.
func NewHandler(handler http.Handler) http.Handler {
	return NewHandlerWith(events.DefaultLogger, handler)
}

// NewHandlerWith wraps the HTTP handler and returns a new handler which logs all
// requests with logger.
//
// Panics from handler are intercepted and trigger a 500 response if no response
// header was sent yet. The panic is not slienced tho and is propagated to the
// parent handler.
func NewHandlerWith(logger *events.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var laddr string

		if value, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			laddr = value.String()
		}

		w := &responseWriter{
			ResponseWriter: res,
			// We capture all the values we need from req in case the object
			// gets modified by the handler.
			logger:  logger,
			request: makeRequest(req, laddr),
		}

		// If the handler panics we want to make sure we report the issue in the
		// access log, while also ensuring that a response is going to be sent
		// down to the client.
		// We don't silence the panic here tho and instead we forward it back to
		// the parent handler which may need to be aware that a panic occurred.
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
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

type responseWriter struct {
	http.ResponseWriter
	logger *events.Logger
	request
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(status int) {
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
		w.request.log(logger, depth+1)
	}
}
