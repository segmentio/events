package httpevents

import (
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events/v2"
)

func BenchmarkRequestLog(b *testing.B) {
	r := newRequest()
	r.laddr = "127.0.0.1:23456"
	r.raddr = "127.0.0.1:80"
	r.method = "GET"
	r.host = "localhost"
	r.path = "/hello/world"
	r.query = "answer=42"
	r.status = 400
	r.statusText = "Bad Request"
	r.loggerMask = LoggerMaskAll

	l := &events.Logger{}

	for i := 0; i != b.N; i++ {
		r.log(l, nil, 0)
	}
}

func equalEvents(e1 events.Event, e2 events.Event) bool {
	// Some fields have unpredicatable values, don't compare them.
	e1.Source = ""
	e2.Source = ""
	e1.Time = time.Time{}
	e2.Time = time.Time{}
	return reflect.DeepEqual(e1, e2)
}
