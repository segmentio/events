package httpevents

import (
	"net/http"

	"github.com/segmentio/events"
)

// NewTransportWith wraps roundTripper and returns a new transport which logs
// all submitted requests with the default logger.
func NewTransport(roundTripper http.RoundTripper) http.RoundTripper {
	return NewTransportWith(events.DefaultLogger, roundTripper)
}

// NewTransportWith wraps roundTripper and returns a new transport which logs
// all submitted requests with logger.
func NewTransportWith(logger *events.Logger, roundTripper http.RoundTripper) http.RoundTripper {
	return &transport{roundTripper, logger}
}

type transport struct {
	http.RoundTripper
	logger *events.Logger
}

func (t *transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	if res, err = t.RoundTripper.RoundTrip(req); res != nil {
		r := acquireRequest(req, "*")
		r.status = res.StatusCode
		r.statusText = http.StatusText(res.StatusCode)
		r.log(t.logger, 1)
		releaseRequest(r)
	}
	return
}
