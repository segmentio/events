package httpevents

import (
	"net/http"

	"github.com/segmentio/events"
)

type request struct {
	laddr    string
	raddr    string
	method   string
	host     string
	path     string
	query    string
	fragment string
	agent    string
	status   int
}

func makeRequest(req *http.Request, laddr string) request {
	var raddr string

	if len(req.RemoteAddr) == 0 {
		raddr = req.URL.Host
	} else {
		raddr = req.RemoteAddr
	}

	return request{
		laddr:    laddr,
		raddr:    raddr,
		method:   req.Method,
		host:     req.Host,
		path:     req.URL.Path,
		query:    req.URL.RawQuery,
		fragment: req.URL.Fragment,
		agent:    req.UserAgent(),
	}
}

func (r *request) log(logger *events.Logger, depth int) {
	var buf [128]byte
	var fmt = append(buf[:0], "%{local_address}s->%{remote_address}s - %{host}s - %{method}s %{path}s"...)

	// Don't output a '?' character when the query string is empty, this is
	// a more natural way of reading URLs.
	if len(r.query) != 0 {
		fmt = append(fmt, '?')
	}
	fmt = append(fmt, "%{query}s"...)

	// Same than with the query string, don't output a '#' character when
	// there is no fragment.
	if len(r.fragment) != 0 {
		fmt = append(fmt, '#')
	}
	fmt = append(fmt, "%{fragment}s - %{status}d %s - %{agent}q"...)

	// Adjust the call depth so we can track the caller of the handler or the
	// transport outside of the httpevents package.
	l := *logger
	l.CallDepth += depth + 1
	l.Log(string(fmt),
		r.laddr,
		r.raddr,
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
