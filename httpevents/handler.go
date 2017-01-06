package httpevents

import (
	"net"
	"net/http"
	"sync"

	"github.com/segmentio/events"
)

// NewHandler wraps the HTTP handler with and access logger which generates
// events for all incoming requests using logger.
func NewHandler(logger *events.Logger, handler http.Handler) http.Handler {
	if logger == nil {
		logger = events.DefaultLogger
	}
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		w := &responseWriter{
			ResponseWriter: res,
			// We capture all the values we need from req in case the object
			// gets modified by the handler.
			logger:  logger,
			address: req.RemoteAddr,
			method:  req.Method,
			host:    req.Host,
			path:    req.URL.Path,
			query:   req.URL.RawQuery,
			agent:   req.UserAgent(),
		}

		// Strip out the port number from the client address, there's little
		// value in knowing which port the request came from since they are most
		// likely chosen randomly by the OS.
		if addr, _, _ := net.SplitHostPort(w.address); len(addr) != 0 {
			w.address = addr
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
	logger  *events.Logger
	address string
	method  string
	host    string
	path    string
	query   string
	agent   string
}

func (w *responseWriter) WriteHeader(status int) {
	if w.logger != nil {
		// 127.0.0.1:43243 -
		w.logger.Log("%:address:s - %:host:s - %:method:s %:path:s?%:query:s - %:status:d %s - %:agent:s",
			w.address,
			w.host,
			w.method,
			w.path,
			w.query,
			status,
			http.StatusText(status),
			w.agent,
		)
		w.logger = nil
	}
	w.ResponseWriter.WriteHeader(status)
}

var responseWriterPool = sync.Pool{
	New: func() interface{} { return &responseWriter{} },
}
