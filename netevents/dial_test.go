package netevents

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events/v2"
)

func TestDialFunc(t *testing.T) {
	evList := []*events.Event{}
	logger := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	dial := DialFuncWith(logger, func(network string, address string) (net.Conn, error) {
		return mockConn{
			laddr: mockAddr{"127.0.0.1:56789", network},
			raddr: mockAddr{address, network},
		}, nil
	})

	conn, err := dial("tcp", "127.0.0.1:80")
	if err != nil {
		t.Error(err)
		return
	}
	conn.Close()

	if len(evList) != 2 {
		t.Error("bad event count:", len(evList))
		return
	}

	e1 := *evList[0]
	e2 := *evList[1]

	t.Logf("%#v", e1)
	t.Logf("%#v", e2)

	e1.Source = ""
	e2.Source = ""
	e1.Time = time.Time{}
	e2.Time = time.Time{}

	if !reflect.DeepEqual(e1, events.Event{
		Message: "127.0.0.1:56789->127.0.0.1:80 - opening client tcp connection",
		Args: events.Args{
			{Name: "local_address", Value: "127.0.0.1:56789"},
			{Name: "remote_address", Value: "127.0.0.1:80"},
			{Name: "event", Value: "opening"},
			{Name: "type", Value: "client"},
			{Name: "protocol", Value: "tcp"},
		},
	}) {
		t.Error("bad opening event")
	}

	if !reflect.DeepEqual(e2, events.Event{
		Message: "127.0.0.1:56789->127.0.0.1:80 - closing client tcp connection",
		Args: events.Args{
			{Name: "local_address", Value: "127.0.0.1:56789"},
			{Name: "remote_address", Value: "127.0.0.1:80"},
			{Name: "event", Value: "closing"},
			{Name: "type", Value: "client"},
			{Name: "protocol", Value: "tcp"},
		},
	}) {
		t.Error("bad closing event")
	}
}
