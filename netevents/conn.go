package netevents

import (
	"net"
	"sync"

	"github.com/segmentio/events"
)

type connLogger struct {
	net.Conn
	*events.Logger
	typ  string
	once sync.Once
}

func (conn *connLogger) BaseConn() net.Conn {
	return conn.Conn
}

func (conn *connLogger) Close() error {
	conn.close(1)
	return conn.Conn.Close()
}

func (conn *connLogger) open(depth int) {
	conn.log(depth+1, "opening")
}

func (conn *connLogger) close(depth int) {
	conn.once.Do(func() { conn.log(depth+3, "closing") })
}

func (conn *connLogger) log(depth int, event string) {
	raddr := conn.RemoteAddr()
	laddr := conn.LocalAddr()

	// Here we're doing our best to adjust the call depth to point outside
	// of this package, but if the connection is wrapped again it may not be
	// very meaningful to report the file/line from which the connection was
	// handled or closed.
	logger := *conn.Logger
	logger.CallDepth += depth + 1

	logger.Log("%{local_address}s->%{remote_address}s - %{event}s %{type}s %{protocol}s connection",
		laddr.String(),
		raddr.String(),
		event,
		conn.typ,
		laddr.Network(),
	)
}
