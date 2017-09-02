package httpevents

import (
	"net/http"
	"sort"
	"sync"

	"github.com/segmentio/events"
)

// The request type is used to generate the log lines of http requests sent by
// a http.RoundTripper or received by a http.Handler.
//
// The package manages a pool of request values to avoid having to reallocate a
// new request object for each request being logged.
type request struct {
	laddr      string
	raddr      string
	method     string
	host       string
	path       string
	query      string
	fragment   string
	agent      string
	headers    headerList
	status     int
	statusText string
	fmtbuf     []byte
	logbuf     []byte
	argbuf     []interface{}
}

func acquireRequest(req *http.Request, laddr string) *request {
	r := requestPool.Get().(*request)
	r.reset(req, laddr)
	return r
}

func releaseRequest(r *request) {
	r.release()
	requestPool.Put(r)
}

func (r *request) release() {
	const zero = ""

	// Don't retain pointers so the garbage collector is free to release them.
	r.laddr = zero
	r.raddr = zero
	r.method = zero
	r.host = zero
	r.path = zero
	r.query = zero
	r.fragment = zero
	r.agent = zero
	r.status = 0
	r.statusText = zero

	for i := range r.headers {
		r.headers[i] = header{}
	}

	for i := range r.argbuf {
		r.argbuf[i] = nil
	}
}

func (r *request) reset(req *http.Request, laddr string) {
	var raddr string

	if len(laddr) == 0 {
		laddr = "???"
	}

	if raddr = req.RemoteAddr; len(raddr) == 0 {
		raddr = "???"
	}

	r.laddr = laddr
	r.raddr = raddr
	r.method = req.Method
	r.host = req.Host
	r.path = req.URL.Path
	r.query = req.URL.RawQuery
	r.fragment = req.URL.Fragment
	r.agent = req.UserAgent()

	if r.headers == nil {
		r.headers = make(headerList, 0, len(req.Header))
	} else {
		r.headers = r.headers[:0]
	}

	for name, values := range req.Header {
		for _, value := range values {
			r.headers = append(r.headers, header{
				name:  name,
				value: value,
			})
		}
	}

	sort.Sort(r)
}

// Implement sort.Interface to sort the list of headers without requiring an
// extra malloc when converting to an interface.
func (r *request) Len() int               { return len(r.headers) }
func (r *request) Less(i int, j int) bool { return r.headers[i].name < r.headers[j].name }
func (r *request) Swap(i int, j int)      { r.headers[i], r.headers[j] = r.headers[j], r.headers[i] }

func (r *request) log(logger *events.Logger, depth int) {
	arg := append(r.argbuf[:0], convS2E(&r.laddr), convS2E(&r.raddr), convS2E(&r.host), convS2E(&r.method))
	fmt := append(r.logbuf[:0], "%{local_address}s->%{remote_address}s - %{host}s - %{method}s"...)

	// Some methods don't have a path (like CONNECT), strip it to avoid printing
	// a double-space.
	if len(r.path) != 0 {
		fmt = append(fmt, " %{path}s"...)
		arg = append(arg, convS2E(&r.path))
	}

	// Don't output a '?' character when the query string is empty, this is
	// a more natural way of reading URLs.
	if len(r.query) != 0 {
		fmt = append(fmt, "?%{query}s"...)
		arg = append(arg, convS2E(&r.query))
	}

	// Same than with the query string, don't output a '#' character when
	// there is no fragment.
	if len(r.fragment) != 0 {
		fmt = append(fmt, "#%{fragment}s"...)
		arg = append(arg, convS2E(&r.fragment))
	}
	fmt = append(fmt, " - %{status}d %s - %q"...)
	arg = append(arg, convI2E(&r.status), convS2E(&r.statusText), convS2E(&r.agent))
	arg = append(arg, events.Args{{
		Name:  "headers",
		Value: &r.headers,
	}})

	// Adjust the call depth so we can track the caller of the handler or the
	// transport outside of the httpevents package.
	l := *logger
	l.CallDepth += depth + 1

	switch {
	case is4xx(r.status) || is5xx(r.status):
		l.Log(bytesToStringNonEmpty(fmt), arg...)
	default:
		l.Debug(bytesToStringNonEmpty(fmt), arg...)
	}

	r.argbuf = arg
	r.logbuf = fmt
}

var requestPool = sync.Pool{
	New: func() interface{} { return newRequest() },
}

func newRequest() *request {
	return &request{
		fmtbuf: make([]byte, 0, 64),
		logbuf: make([]byte, 0, 128),
		argbuf: make([]interface{}, 0, 10),
	}
}

func is4xx(status int) bool {
	return status >= 400 && status <= 499
}

func is5xx(status int) bool {
	return status >= 500 && status <= 599
}
