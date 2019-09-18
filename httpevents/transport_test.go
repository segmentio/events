package httpevents

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/events/v2"
	"github.com/segmentio/events/v2/eventstest"
)

func TestTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Date", "today")
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
	defer res.Body.Close()

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
			{Name: "local_address", Value: "*"},
			{Name: "remote_address", Value: "192.0.2.1:1234"},
			{Name: "host", Value: req.Host},
			{Name: "method", Value: "GET"},
			{Name: "path", Value: "/"},
			{Name: "status", Value: 200},
			{Name: "request", Value: &headerList{
				{name: "User-Agent", value: "httpevents"},
			}},
			{Name: "response", Value: &headerList{
				{name: "Content-Length", value: "12"},
				{name: "Content-Type", value: "text/plain; charset=utf-8"},
				{name: "Date", value: "today"},
			}},
		},
		Debug: true,
	})
}
