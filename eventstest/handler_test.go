package eventstest

import (
	"testing"

	"github.com/segmentio/events"
)

func TestHandler(t *testing.T) {
	h := &Handler{}
	logger := events.NewLogger(h)

	logger.Log("Hello %{name}s!", "Luke", events.Args{
		{"from", "Han"},
	})

	h.AssertEvents(t, events.Event{
		Message: "Hello Luke!",
		Args: events.Args{
			{"name", "Luke"},
			{"from", "Han"},
		},
	})
}
