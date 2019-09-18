package netevents

import (
	"context"
	"net"

	"github.com/segmentio/events/v2"
)

// Handler is an interface that can be implemented by types that serve network
// connections.
type Handler interface {
	ServeConn(ctx context.Context, conn net.Conn)
}

// NewHandler returns a wrapped handler which logs with the default logger
// all opening and closing events that occur on the connections servied by the
// handler.
func NewHandler(handler Handler) Handler {
	return NewHandlerWith(events.DefaultLogger, handler)
}

// NewHandlerWith returns a wrapped handler which logs with logger all
// opening and closing events that occur on the connections served by the
// handler with logger.
func NewHandlerWith(logger *events.Logger, handler Handler) Handler {
	if logger == nil {
		panic("cannot send network events to a nil logger")
	}

	if handler == nil {
		panic("cannot capture network events of a nil handler")
	}

	return &handlerLogger{
		Handler: handler,
		Logger:  logger,
	}
}

type handlerLogger struct {
	Handler
	*events.Logger
}

func (h *handlerLogger) ServeConn(ctx context.Context, conn net.Conn) {
	c := &connLogger{Conn: conn, Logger: h.Logger, typ: "server"}
	c.open(2)
	defer c.close(2) // ensure we always log the close event
	h.Handler.ServeConn(ctx, c)
}
