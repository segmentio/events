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
		laddr: mockAddr{"127.0.0.1:80", "tcp"},
		raddr: mockAddr{"127.0.0.1:56789", "tcp"},
	})

	e1 := events.Event{
		Message: "127.0.0.1:80->127.0.0.1:56789 - opening server tcp connection",
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"event", "opening"},
			{"type", "server"},
			{"protocol", "tcp"},
		},
	}

	e2 := events.Event{
		Message: "127.0.0.1:80->127.0.0.1:56789 - closing server tcp connection",
		Args: events.Args{
			{"local_address", "127.0.0.1:80"},
			{"remote_address", "127.0.0.1:56789"},
			{"event", "closing"},
			{"type", "server"},
			{"protocol", "tcp"},
		},
	}

	h.AssertEvents(t, e1, e2)
}
