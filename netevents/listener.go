package netevents

import (
	"net"
	"sync/atomic"

	"github.com/segmentio/events"
	"github.com/segmentio/netx"
)

// NewListener returns a wrapped version of lstn which logs with the default
// logger all accept and close events that occur on the listener.
func NewListener(lstn net.Listener) net.Listener {
	return NewListenerWith(events.DefaultLogger, lstn)
}

// NewListenerWith returns a wrapped version of lstn which logs with logger
// all accept and close events that occur on the listener.
func NewListenerWith(logger *events.Logger, lstn net.Listener) net.Listener {
	return &listener{
		lstn:   lstn,
		logger: logger,
	}
}

type listener struct {
	lstn   net.Listener
	logger *events.Logger
	closed uint32
}

func (l *listener) Accept() (conn net.Conn, err error) {
	if conn, err = l.lstn.Accept(); err != nil {
		if atomic.LoadUint32(&l.closed) == 0 {
			l.error("accept", err)
		}
	}

	if conn != nil {
		cl := &connLogger{
			Conn:   conn,
			Logger: l.logger,
			typ:    "server",
		}
		cl.open(1)
		conn = cl
	}

	return
}

func (l *listener) Close() (err error) {
	if atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		addr := l.Addr()
		logger := *l.logger
		logger.CallDepth++
		logger.Debug("%{local_address}s - %{event}s %{type}s %{protocol}s socket",
			addr.String(), "shutting down", "server", addr.Network(),
		)
	}
	return l.lstn.Close()
}

func (l *listener) Addr() net.Addr {
	return l.lstn.Addr()
}

func (l *listener) error(op string, err error) {
	logger := *l.logger
	logger.CallDepth += 2

	if netx.IsTemporary(err) {
		logger.Debug("%{local_address}s - temporary error accepting connections - %{error}s",
			l.Addr(), err,
		)
	} else {
		logger.Log("%{local_address}s - error accepting connection - %{error}s",
			l.Addr(), err,
		)
	}
}
