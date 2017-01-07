package httpevents

import (
	"net/http"

	"github.com/segmentio/events"
)

// NewTransport wraps roundTripper and returns a new transport which logs all
// submitted requests with logger.
func NewTransport(logger *events.Logger, roundTripper http.RoundTripper) http.RoundTripper {
	if logger == nil {
		logger = events.DefaultLogger
	}
	return &transport{roundTripper, logger}
}

type transport struct {
	http.RoundTripper
	logger *events.Logger
}

func (t *transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	r := makeRequest(req, "*")

	if res, err = t.RoundTripper.RoundTrip(req); res != nil {
		r.status = res.StatusCode
		r.log(t.logger, 1)
	}

	return
}
