package netevents

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events"
)

func TestListener(t *testing.T) {
	evList := []*events.Event{}
	logger := events.NewLogger(events.HandlerFunc(func(e *events.Event) {
		evList = append(evList, e.Clone())
	}))

	l, err := net.Listen("tcp", ":0")
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
		&events.Event{
			Message: fmt.Sprintf("%s->%s - opening server tcp connection", laddr, raddr),
			Args: events.Args{
				{"local_address", laddr},
				{"remote_address", raddr},
				{"event", "opening"},
				{"type", "server"},
				{"protocol", "tcp"},
			},
		},

		&events.Event{
			Message: fmt.Sprintf("%s->%s - closing server tcp connection", laddr, raddr),
			Args: events.Args{
				{"local_address", laddr},
				{"remote_address", raddr},
				{"event", "closing"},
				{"type", "server"},
				{"protocol", "tcp"},
			},
		},

		&events.Event{
			Message: fmt.Sprintf("%s - shutting down server tcp socket", saddr),
			Args: events.Args{
				{"local_address", saddr},
				{"event", "shutting down"},
				{"type", "server"},
				{"protocol", "tcp"},
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
