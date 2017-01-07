package netevents

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events"
	"github.com/segmentio/netx"
)

func TestHandler(t *testing.T) {
	evList := []*events.Event{}
	logger := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	handler := NewHandler(logger, netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		conn.Close()
	}))

	handler.ServeConn(context.Background(), mockConn{
		laddr: mockAddr{"127.0.0.1:80", "tcp"},
		raddr: mockAddr{"127.0.0.1:56789", "tcp"},
	})

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
		Message: "127.0.0.1:80->127.0.0.1:56789 - opening server tcp connection",
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"event", "opening"},
			{"type", "server"},
			{"protocol", "tcp"},
		},
	}) {
		t.Error("bad opening event")
	}

	if !reflect.DeepEqual(e2, events.Event{
		Message: "127.0.0.1:80->127.0.0.1:56789 - closing server tcp connection",
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"event", "closing"},
			{"type", "server"},
			{"protocol", "tcp"},
		},
	}) {
		t.Error("bad closing event")
	}
}
