package eventstest

import (
	"reflect"
	"sync"
	"testing"

	"github.com/segmentio/events/v2"
)

var _ events.Handler = (*Handler)(nil)

// Handler is a stats handler that can record events for inspection and make
// assertions on them.
type Handler struct {
	evList []events.Event
	sync.Mutex
}

// HandleEvent implements the events.Handler interface.
func (h *Handler) HandleEvent(e *events.Event) {
	h.Lock()
	defer h.Unlock()
	h.evList = append(h.evList, *e.Clone())
}

// AssertEvents asserts that the given events were handled by the handler.
// It only checks the following event fields for equality — .Args, .Debug and .Message.
// It ignores the .Source and .Time fields.
func (h *Handler) AssertEvents(t testing.TB, expectedEvents ...events.Event) {
	h.Lock()
	defer h.Unlock()

	if len(h.evList) != len(expectedEvents) {
		t.Errorf("expected %d events but got %d events: %v", len(expectedEvents), len(h.evList), h.evList)
		return
	}

	for i := 0; i < len(h.evList); i++ {
		got := h.evList[i]
		expected := expectedEvents[i]
		if !assertEqualEvent(t, got, expected) {
			t.Error("bad events at index", i)
			t.Log("expected =>", expected)
			t.Log("got ======>", got)
			return
		}
	}
}

func assertEqualEvent(t testing.TB, got, expected events.Event) bool {
	return assertEqualField(t, "Args", got.Args, expected.Args) &&
		assertEqualField(t, "Debug", got.Debug, expected.Debug) &&
		assertEqualField(t, "Message", got.Message, expected.Message)
}

func assertEqualField(t testing.TB, field string, got, expected interface{}) bool {
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("bad .%s:", field)
		t.Log("expected =>", expected)
		t.Log("got ======>", got)
		return false
	}
	return true
}
