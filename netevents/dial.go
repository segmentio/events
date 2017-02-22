package netevents

import (
	"context"
	"net"
	"runtime"

	"github.com/segmentio/events"
)

// DialFunc returns a wrapper for dial which logs with the default logger all
// opening and closing events that occur on the connections returned by the dial
// function.
func DialFunc(dial func(string, string) (net.Conn, error)) func(string, string) (net.Conn, error) {
	return DialFuncWith(events.DefaultLogger, dial)
}

// DialFuncWith returns a wrapper for dial which logs with logger all opening
// and closing events that occur on the connections returned by the dial
// function.
func DialFuncWith(logger *events.Logger, dial func(string, string) (net.Conn, error)) func(string, string) (net.Conn, error) {
	dialContext := dialContextFunc(1, logger,
		func(ctx context.Context, network string, address string) (net.Conn, error) {
			return dial(network, address)
		},
	)
	return func(network string, address string) (net.Conn, error) {
		return dialContext(context.Background(), network, address)
	}
}

// DialContextFuncWith returns a wrapper for dial which logs with the default
// logger all opening and closing events that occur on the connections returned
// by the dial function.
func DialContextFunc(dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return DialContextFuncWith(events.DefaultLogger, dial)
}

// DialContextFuncWith returns a wrapper for dial which logs with logger all
// opening and closing events that occur on the connections returned by the
// dial function.
func DialContextFuncWith(logger *events.Logger, dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return dialContextFunc(0, logger, dial)
}

func dialContextFunc(depth int, logger *events.Logger, dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network string, address string) (conn net.Conn, err error) {
		if conn, err = dial(ctx, network, address); err == nil {
			c := &connLogger{Conn: conn, Logger: logger, typ: "client"}
			c.open(depth + 1)
			conn = c
			// Ensure the connection emits a close event if it's closed by the
			// garbage collector.
			runtime.SetFinalizer(c, func(c *connLogger) { c.close(0) })
		}
		return
	}
}
