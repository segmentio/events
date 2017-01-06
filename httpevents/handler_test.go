package httpevents

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events"
)

func TestHandler(t *testing.T) {
	evList := []*events.Event{}

	req := httptest.NewRequest("GET", "/hello?answer=42", nil)
	req.Header.Set("User-Agent", "httpevents")
	req.Host = "www.github.com"
	req.RemoteAddr = "127.0.0.1"

	res := httptest.NewRecorder()
	log := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	h := NewHandler(log, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusAccepted)
	}))
	h.ServeHTTP(res, req)

	if len(evList) != 1 {
		t.Error("bad event count:", len(evList))
		return
	}

	e := *evList[0]
	e.Source = ""
	e.Time = time.Time{}

	_ = events.Event{
		Message: "127.0.0.1 - www.github.com - GET /hello?answer=42 - 202 Accepted - httpevents",
		Source:  "",
		Args: events.Args{
			events.Arg{Name: "address", Value: "127.0.0.1"},
			events.Arg{Name: "host", Value: "www.github.com"},
			events.Arg{Name: "method", Value: "GET"},
			events.Arg{Name: "path", Value: "/hello"},
			events.Arg{Name: "query", Value: "answer=42"},
			events.Arg{Name: "status", Value: 202},
			events.Arg{Name: "", Value: "Accepted"},
			events.Arg{Name: "agent", Value: "httpevents"},
		},
		Time:  time.Time{},
		Debug: true,
	}

	if !reflect.DeepEqual(e, events.Event{
		Message: "127.0.0.1 - www.github.com - GET /hello?answer=42 - 201 Accepted - httpevents",
		Args: events.Args{
			{"address", "127.0.0.1"},
			{"host", "www.github.com"},
			{"method", "GET"},
			{"path", "/hello"},
			{"query", "answer=42"},
			{"status", 202},
			{"agent", "httpevents"},
		},
	}) {
		t.Errorf("%#v", e)
	}
}
