package netevents

import (
	"context"
	"net"
	"testing"

	"github.com/segmentio/events"
	"github.com/segmentio/events/eventstest"
)

type testHandler struct{}

func (*testHandler) ServeConn(ctx context.Context, conn net.Conn) {
	conn.Close()
}

func TestHandler(t *testing.T) {
	h := &eventstest.Handler{}
	logger := events.NewLogger(h)

	handler := NewHandlerWith(logger, &testHandler{})

	handler.ServeConn(context.Background(), mockConn{
		laddr: mockAddr{s: "127.0.0.1:80", n: "tcp"},
		raddr: mockAddr{s: "127.0.0.1:56789", n: "tcp"},
	})

	e1 := events.Event{
		Message: "127.0.0.1:80->127.0.0.1:56789 - opening server tcp connection",
		Args: events.Args{
			{Name: "local_address", Value: "127.0.0.1:80"},
			{Name: "remote_address", Value: "127.0.0.1:56789"},
			{Name: "event", Value: "opening"},
			{Name: "type", Value: "server"},
			{Name: "protocol", Value: "tcp"},
		},
	}

	e2 := events.Event{
		Message: "127.0.0.1:80->127.0.0.1:56789 - closing server tcp connection",
		Args: events.Args{
			{Name: "local_address", Value: "127.0.0.1:80"},
			{Name: "remote_address", Value: "127.0.0.1:56789"},
			{Name: "event", Value: "closing"},
			{Name: "type", Value: "server"},
			{Name: "protocol", Value: "tcp"},
		},
	}

	h.AssertEvents(t, e1, e2)
}
