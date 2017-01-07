package httpevents

import (
	"net/http"

	"github.com/segmentio/events"
)

// NewTransport wraps the HTTP transport and returns a new RoundTripper which
// logs all submitted requests with logger.
func NewTransport(logger *events.Logger, transport http.RoundTripper) http.RoundTripper {
	if logger == nil {
		logger = events.DefaultLogger
	}
	return roundTripperFunc(func(req *http.Request) (res *http.Response, err error) {
		//r := makeRequest(req)

		if res, err = transport.RoundTrip(req); err != nil {

		} else {

		}

		return
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
