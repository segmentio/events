package netevents

import (
	"context"
	"net"

	"github.com/segmentio/events"
	"github.com/segmentio/netx"
)

// NewHandler returns a wrapper for handler which logs opening and closing
// events that occur on the connections served by the handler with logger.
//
// The logger may be nil, in that case the default logger is used to report
// the connection events.
func NewHandler(logger *events.Logger, handler netx.Handler) netx.Handler {
	if logger == nil {
		logger = events.DefaultLogger
	}
	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		c := &connLogger{Conn: conn, Logger: logger, typ: "server"}
		c.open(2)
		defer c.close(2) // ensure we always log the close event
		handler.ServeConn(ctx, c)
	})
}
