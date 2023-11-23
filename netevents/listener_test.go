package netevents

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events/v2"
)

func TestListener(t *testing.T) {
	evList := []*events.Event{}
	logger := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	l = NewListenerWith(logger, l)

	c, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	a, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}

	saddr := l.Addr().String()
	laddr := a.LocalAddr().String()
	raddr := a.RemoteAddr().String()

	a.Close()
	c.Close()
	l.Close()

	for _, e := range evList {
		e.Source = ""
		e.Time = time.Time{}
	}

	expect := []*events.Event{
		{
			Message: fmt.Sprintf("%s->%s - opening server tcp connection", laddr, raddr),
			Args: events.Args{
				{Name: "local_address", Value: laddr},
				{Name: "remote_address", Value: raddr},
				{Name: "event", Value: "opening"},
				{Name: "type", Value: "server"},
				{Name: "protocol", Value: "tcp"},
			},
		},

		{
			Message: fmt.Sprintf("%s->%s - closing server tcp connection", laddr, raddr),
			Args: events.Args{
				{Name: "local_address", Value: laddr},
				{Name: "remote_address", Value: raddr},
				{Name: "event", Value: "closing"},
				{Name: "type", Value: "server"},
				{Name: "protocol", Value: "tcp"},
			},
		},

		{
			Message: fmt.Sprintf("%s - shutting down server tcp socket", saddr),
			Args: events.Args{
				{Name: "local_address", Value: saddr},
				{Name: "event", Value: "shutting down"},
				{Name: "type", Value: "server"},
				{Name: "protocol", Value: "tcp"},
			},
			Debug: true,
		},
	}

	if !reflect.DeepEqual(evList, expect) {
		t.Error("bad event list:")
		t.Log("expected:")
		for _, e := range expect {
			t.Logf("%#v\n", e)
		}
		t.Log("found:")
		for _, e := range evList {
			t.Logf("%#v\n", e)
		}
	}
}
