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

	if len(laddr) == 0 {
		laddr = "???"
	}

	if raddr = req.RemoteAddr; len(raddr) == 0 {
		raddr = "???"
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
	var mem [10]interface{}
	var arg = append(mem[:0], r.laddr, r.raddr, r.host, r.method)

	var buf [128]byte
	var fmt = append(buf[:0], "%{local_address}s->%{remote_address}s - %{host}s - %{method}s"...)

	// Some methods don't have a path (like CONNECT), strip it to avoid printing
	// a double-space.
	if len(r.path) != 0 {
		fmt = append(fmt, " %{path}s"...)
		arg = append(arg, r.path)
	}

	// Don't output a '?' character when the query string is empty, this is
	// a more natural way of reading URLs.
	if len(r.query) != 0 {
		fmt = append(fmt, "?%{query}s"...)
		arg = append(arg, r.query)
	}

	// Same than with the query string, don't output a '#' character when
	// there is no fragment.
	if len(r.fragment) != 0 {
		fmt = append(fmt, "#%{fragment}s"...)
		arg = append(arg, r.fragment)
	}
	fmt = append(fmt, " - %{status}d %s - %{agent}q"...)
	arg = append(arg, r.status, http.StatusText(r.status), r.agent)

	// Adjust the call depth so we can track the caller of the handler or the
	// transport outside of the httpevents package.
	l := *logger
	l.CallDepth += depth + 1
	l.Log(string(fmt), arg...)
}
