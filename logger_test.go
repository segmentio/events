package events

import (
	"reflect"
	"testing"
	"time"
)

var appendFormatTests = []struct {
	srcFmt  string
	srcArgs []interface{}
	dstFmt  string
	dstArgs Args
}{
	{ // empty format
		srcFmt:  "",
		srcArgs: nil,
		dstFmt:  "",
		dstArgs: nil,
	},
	{ // simple string
		srcFmt:  "Hello World!",
		srcArgs: nil,
		dstFmt:  "Hello World!",
		dstArgs: nil,
	},
	{ // fmt-like format
		srcFmt:  "Hello %s!",
		srcArgs: []interface{}{"Luke"},
		dstFmt:  "Hello %s!",
		dstArgs: nil,
	},
	{ // simple format
		srcFmt:  "Hello %{name}s!",
		srcArgs: []interface{}{"Luke"},
		dstFmt:  "Hello %s!",
		dstArgs: Args{{"name", "Luke"}},
	},
	{ // complex format
		srcFmt:  "{ %{first-name}q: %{last-name}#v }",
		srcArgs: []interface{}{"Luke", "Skywalker"},
		dstFmt:  "{ %q: %#v }",
		dstArgs: Args{{"first-name", "Luke"}, {"last-name", "Skywalker"}},
	},
	{ // escaped format
		srcFmt:  "%%{",
		srcArgs: nil,
		dstFmt:  "%%{",
		dstArgs: nil,
	},
	{ // trailing ':'
		srcFmt:  "%{",
		srcArgs: nil,
		dstFmt:  "%{",
		dstArgs: nil,
	},
	{ // unclosed ':'
		srcFmt:  "%{name",
		srcArgs: nil,
		dstFmt:  "%{name",
		dstArgs: nil,
	},
	{ // missing arg
		srcFmt:  "Hello %{name}s",
		srcArgs: nil,
		dstFmt:  "Hello %s",
		dstArgs: Args{{"name", "MISSING"}},
	},
}

func TestAppendFormat(t *testing.T) {
	for _, test := range appendFormatTests {
		t.Run(test.srcFmt, func(t *testing.T) {
			dstFmt, dstArgs := appendFormat(nil, nil, test.srcFmt, test.srcArgs)

			if s := string(dstFmt); s != test.dstFmt {
				t.Error("format:", s)
			}

			if !reflect.DeepEqual(dstArgs, test.dstArgs) {
				t.Error("args:", dstArgs)
			}
		})
	}
}

func BenchmarkAppendFormat(b *testing.B) {
	dstFmt := make([]byte, 0, 1024)
	dstArgs := make(Args, 0, 8)

	for _, test := range appendFormatTests {
		b.Run(test.srcFmt, func(b *testing.B) {
			for i := 0; i != b.N; i++ {
				appendFormat(dstFmt[:0], dstArgs[:0], test.srcFmt, test.srcArgs)
			}
		})
	}
}

var loggerTests = []struct {
	format string
	args   []interface{}
	event  Event
}{
	{ /* all zero-values */ },
	{
		format: "Hello %{name}s!",
		args:   []interface{}{"Luke", Args{{"from", "Han"}}},
		event: Event{
			Message: "Hello Luke!",
			Args:    Args{{"name", "Luke"}, {"from", "Han"}},
		},
	},
}

func TestLogger(t *testing.T) {
	events := []*Event{}
	logger := Logger{
		Handler: HandlerFunc(func(e *Event) {
			events = append(events, e.Clone())
		}),
		EnableSource: true,
	}

	for _, test := range loggerTests {
		t.Run(string(test.event.Message), func(t *testing.T) {
			t.Run("Log", func(t *testing.T) {
				events = events[:0]
				logger.EnableDebug = false
				logger.Log(test.format, test.args...)
				checkEvents(t, events, []*Event{&test.event})
			})
			t.Run("Debug", func(t *testing.T) {
				t.Run("enabled", func(t *testing.T) {
					events = events[:0]
					logger.EnableDebug = true
					logger.Debug(test.format, test.args...)

					testEvent := test.event
					testEvent.Debug = true
					checkEvents(t, events, []*Event{&testEvent})
				})
				t.Run("disabled", func(t *testing.T) {
					events = events[:0]
					logger.EnableDebug = false
					logger.Debug(test.format, test.args...)
					checkEvents(t, events, nil)
				})
			})
		})
	}

	t.Run("With", func(t *testing.T) {
		logger.EnableDebug = false

		child1 := logger.With(Args{{"hello", "world"}})
		child2 := child1.With(Args{{"question", "how are you?"}})

		events = events[:0]
		child1.Log("child1")
		child2.Log("child2", Args{{"answer", 42}})

		checkEvents(t, events, []*Event{
			{
				Message: "child1",
				Args:    Args{{"hello", "world"}},
			},
			{
				Message: "child2",
				Args:    Args{{"hello", "world"}, {"question", "how are you?"}, {"answer", 42}},
			},
		})
	})
}

func BenchmarkLogger(b *testing.B) {
	logger := Logger{
		Handler: Discard,
	}

	for _, test := range loggerTests {
		b.Run(string(test.event.Message), func(b *testing.B) {
			for i := 0; i != b.N; i++ {
				logger.Log(test.format, test.args...)
			}
		})
	}

	b.Run("inline", func(b *testing.B) {
		for i := 0; i != b.N; i++ {
			logger.Log("[%d, %d, %d]", 1, 2, 3)
		}
	})
}

func checkEvents(t *testing.T, e1 []*Event, e2 []*Event) {
	if len(e1) != len(e2) {
		t.Error("length mismatch:", len(e1), "!=", len(e2))
		t.Log(e1)
		t.Log(e2)
		return
	}

	for i := range e1 {
		v1 := *e1[i]
		v2 := *e2[i]
		v1.Source = ""
		v2.Source = ""
		v1.Time = time.Time{}
		v2.Time = time.Time{}
		if !reflect.DeepEqual(v1, v2) {
			t.Error("event mismatch at index", i)
			t.Logf("%#v", v1)
			t.Logf("%#v", v2)
		}
	}
}
