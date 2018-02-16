package eventstest

import (
	"testing"

	"github.com/segmentio/events"
)

func TestHandler(t *testing.T) {
	h := &Handler{}
	logger := events.NewLogger(h)

	logger.Log("Hello %{name}s!", "Luke", events.Args{
		{Name: "from", Value: "Han"},
	})

	h.AssertEvents(t, events.Event{
		Message: "Hello Luke!",
		Args: events.Args{
			{Name: "name", Value: "Luke"},
			{Name: "from", Value: "Han"},
		},
	})
}
