package netevents

import (
	"net"
	"time"
)

type mockConn struct {
	laddr mockAddr
	raddr mockAddr
}

func (c mockConn) Close() error { return nil }

func (c mockConn) Read(b []byte) (int, error)  { return 0, nil }
func (c mockConn) Write(b []byte) (int, error) { return len(b), nil }

func (c mockConn) SetDeadline(t time.Time) error      { return nil }
func (c mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c mockConn) SetWriteDeadline(t time.Time) error { return nil }

func (c mockConn) LocalAddr() net.Addr  { return c.laddr }
func (c mockConn) RemoteAddr() net.Addr { return c.raddr }

type mockAddr struct {
	s string
	n string
}

func (a mockAddr) String() string  { return a.s }
func (a mockAddr) Network() string { return a.n }
