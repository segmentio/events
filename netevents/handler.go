package netevents

import (
	"context"
	"net"

	"github.com/segmentio/events"
	"github.com/segmentio/netx"
)

// NewHandler returns a wrapped handler which logs with the default logger
// all opening and closing events that occur on the connections servied by the
// handler.
func NewHandler(handler netx.Handler) netx.Handler {
	return NewHandlerWith(events.DefaultLogger, handler)
}

// NewHandlerWith returns a wrapped handler which logs with logger all
// opening and closing events that occur on the connections served by the
// handler with logger.
func NewHandlerWith(logger *events.Logger, handler netx.Handler) netx.Handler {
	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		c := &connLogger{Conn: conn, Logger: logger, typ: "server"}
		c.open(2)
		defer c.close(2) // ensure we always log the close event
		handler.ServeConn(ctx, c)
	})
}
