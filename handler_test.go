package events

import "testing"

func TestMultiHandler(t *testing.T) {
	n := 0
	m := MultiHandler{
		HandlerFunc(func(e *Event) { n++ }),
		HandlerFunc(func(e *Event) { n++ }),
	}

	m.HandleEvent(&Event{})

	if n != 2 {
		t.Error("bad count of handler received the event:", n)
	}
}
