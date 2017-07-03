package httpevents

import (
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/events"
)

func BenchmarkRequestLog(b *testing.B) {
	b.Run("safe", func(b *testing.B) {
		benchmarkRequestLog(b)
	})

	b.Run("unsafe", func(b *testing.B) {
		events.EnableUnsafeOptimizations = true
		benchmarkRequestLog(b)
		events.EnableUnsafeOptimizations = false
	})
}

func benchmarkRequestLog(b *testing.B) {
	r := newRequest()
	r.laddr = "127.0.0.1:23456"
	r.raddr = "127.0.0.1:80"
	r.method = "GET"
	r.host = "localhost"
	r.path = "/hello/world"
	r.query = "answer=42"
	r.status = 400
	r.statusText = "Bad Request"

	l := &events.Logger{}

	for i := 0; i != b.N; i++ {
		r.log(l, 0)
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
