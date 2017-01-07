package httpevents

import (
	"net/http"

	"github.com/segmentio/events"
)

// NewHandler wraps the HTTP handler and returns a new handler which logs all
// requests to logger.
//
// Panics from handler are intercepted and trigger a 500 response if no response
// header was sent yet. The panic is not slienced tho and is propagated to the
// parent handler.
func NewHandler(logger *events.Logger, handler http.Handler) http.Handler {
	if logger == nil {
		logger = events.DefaultLogger
	}
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		w := &responseWriter{
			ResponseWriter: res,
			// We capture all the values we need from req in case the object
			// gets modified by the handler.
			logger:   logger,
			address:  req.RemoteAddr,
			method:   req.Method,
			host:     req.Host,
			path:     req.URL.Path,
			query:    req.URL.RawQuery,
			fragment: req.URL.Fragment,
			agent:    req.UserAgent(),
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
	logger   *events.Logger
	address  string
	method   string
	host     string
	path     string
	query    string
	fragment string
	agent    string
}

func (w *responseWriter) WriteHeader(status int) {
	if w.logger != nil {
		var buf [128]byte
		var fmt = append(buf[:0], "%{address}s - %{host}s - %{method}s %{path}s"...)

		// Don't output a '?' character when the query string is empty, this is
		// a more natural way of reading URLs.
		if len(w.query) != 0 {
			fmt = append(fmt, '?')
		}
		fmt = append(fmt, "%{query}s"...)

		// Same than with the query string, don't output a '#' character when
		// there is no fragment.
		if len(w.fragment) != 0 {
			fmt = append(fmt, '#')
		}
		fmt = append(fmt, "%{fragment}s - %{status}d %s - %{agent}q"...)

		w.logger.Log(string(fmt),
			w.address,
			w.host,
			w.method,
			w.path,
			w.query,
			w.fragment,
			status,
			http.StatusText(status),
			w.agent,
		)
		w.logger = nil
	}
	w.ResponseWriter.WriteHeader(status)
}
