package netevents

import (
	"context"
	"net"

	"github.com/segmentio/events"
)

// DialFunc returns a wrapper for dial which logs opening and closing events
// that occur on the connections returned by the function.
//
// The logger may be nil, in that case the default logger is used to report
// the connection events.
func DialFunc(logger *events.Logger, dial func(string, string) (net.Conn, error)) func(string, string) (net.Conn, error) {
	dialContext := dialContextFunc(1, logger,
		func(ctx context.Context, network string, address string) (net.Conn, error) {
			return dial(network, address)
		},
	)
	return func(network string, address string) (net.Conn, error) {
		return dialContext(context.Background(), network, address)
	}
}

// DialContextFunc returns a wrapper for dial which logs opening and closing
// events that occur on the connections returned by the function.
//
// The logger may be nil, in that case the default logger is used to report
// the connection events.
func DialContextFunc(logger *events.Logger, dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return dialContextFunc(0, logger, dial)
}

func dialContextFunc(depth int, logger *events.Logger, dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	if logger == nil {
		logger = events.DefaultLogger
	}
	return func(ctx context.Context, network string, address string) (conn net.Conn, err error) {
		if conn, err = dial(ctx, network, address); err == nil {
			c := &connLogger{Conn: conn, Logger: logger, typ: "client"}
			c.open(depth + 1)
			conn = c
		}
		return
	}
}
