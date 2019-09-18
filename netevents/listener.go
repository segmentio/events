package netevents

import (
	"net"
	"sync/atomic"

	"github.com/segmentio/events/v2"
)

// NewListener returns a wrapped version of listener which logs with the default
// logger all accept and close events that occur on the listener.
func NewListener(listener net.Listener) net.Listener {
	return NewListenerWith(events.DefaultLogger, listener)
}

// NewListenerWith returns a wrapped version of listener which logs with logger
// all accept and close events that occur on the listener.
func NewListenerWith(logger *events.Logger, listener net.Listener) net.Listener {
	return &listenerLogger{
		Listener: listener,
		Logger:   logger,
	}
}

type listenerLogger struct {
	net.Listener
	*events.Logger
	closed uint32
}

func (l *listenerLogger) Accept() (conn net.Conn, err error) {
	if conn, err = l.Listener.Accept(); err != nil {
		if atomic.LoadUint32(&l.closed) == 0 {
			l.error("accept", err)
		}
	}

	if conn != nil {
		cl := &connLogger{
			Conn:   conn,
			Logger: l.Logger,
			typ:    "server",
		}
		cl.open(1)
		conn = cl
	}

	return
}

func (l *listenerLogger) Close() (err error) {
	if atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		addr := l.Addr()
		logger := *l.Logger
		logger.CallDepth++
		logger.Debug("%{local_address}s - %{event}s %{type}s %{protocol}s socket",
			addr.String(), "shutting down", "server", addr.Network(),
		)
	}
	return l.Listener.Close()
}

func (l *listenerLogger) error(op string, err error) {
	logger := *l.Logger
	logger.CallDepth += 2

	if isTemporary(err) {
		logger.Debug("%{local_address}s - temporary error accepting connections - %{error}s",
			l.Addr(), err,
		)
	} else {
		logger.Log("%{local_address}s - error accepting connection - %{error}s",
			l.Addr(), err,
		)
	}
}

func isTemporary(err error) bool {
	e, ok := err.(interface {
		Temporary() bool
	})
	return ok && e.Temporary()
}
