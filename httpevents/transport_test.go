package httpevents

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/events"
	"github.com/segmentio/events/eventstest"
)

func TestTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Hello World!"))
	}))
	defer server.Close()

	h := &eventstest.Handler{}
	logger := events.NewLogger(h)

	transport := NewTransportWith(logger, http.DefaultTransport)

	req := httptest.NewRequest("GET", server.URL+"/", nil)
	req.Header.Set("User-Agent", "httpevents")

	res, err := transport.RoundTrip(req)
	if err != nil {
		t.Error(err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}

	if s := string(b); s != "Hello World!" {
		t.Error("bad response:", s)
	}

	h.AssertEvents(t, events.Event{
		Message: fmt.Sprintf(`*->192.0.2.1:1234 - %s - GET / - 200 OK - "httpevents"`, req.Host),
		Args: events.Args{
			{"local_address", "*"},
			{"remote_address", "192.0.2.1:1234"},
			{"host", req.Host},
			{"method", "GET"},
			{"path", "/"},
			{"status", 200},
			{"headers", &headerList{{"User-Agent", "httpevents"}}},
		},
		Debug: true,
	})
}
