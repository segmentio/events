package httpevents

import (
	"net/http"

	"github.com/segmentio/events"
)

type request struct {
	address  string
	method   string
	host     string
	path     string
	query    string
	fragment string
	agent    string
	status   int
}

func makeRequest(req *http.Request) request {
	return request{
		address:  req.RemoteAddr,
		method:   req.Method,
		host:     req.Host,
		path:     req.URL.Path,
		query:    req.URL.RawQuery,
		fragment: req.URL.Fragment,
		agent:    req.UserAgent(),
	}
}

func (r *request) log(logger *events.Logger, depth int) {
	var l = *logger
	var b [128]byte

	l.CallDepth = depth + 1
	l.Log(string(r.appendLogFormat(b[:0])),
		r.address,
		r.host,
		r.method,
		r.path,
		r.query,
		r.fragment,
		r.status,
		http.StatusText(r.status),
		r.agent,
	)
}

func (r *request) appendLogFormat(b []byte) []byte {
	b = append(b, "%{address}s - %{host}s - %{method}s %{path}s"...)

	// Don't output a '?' character when the query string is empty, this is
	// a more natural way of reading URLs.
	if len(r.query) != 0 {
		b = append(b, '?')
	}
	b = append(b, "%{query}s"...)

	// Same than with the query string, don't output a '#' character when
	// there is no fragment.
	if len(r.fragment) != 0 {
		b = append(b, '#')
	}
	return append(b, "%{fragment}s - %{status}d %s - %{agent}q"...)
}
