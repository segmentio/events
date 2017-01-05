package events

// The Handler interface is implemented by types that intend to be event routers
// or apply transformations to an event before forwarding it to another handler.
type Handler interface {
	// HandleEvent is called by event producers on their event handler, passing
	// one event object as argument to the function.
	//
	// The handler MUST NOT retain any references to the event or its fields. If
	// the handler needs to capture the event is has to create a copy by calling
	// e.Clone.
	HandleEvent(e *Event)
}

// HandlerFunc makes it possible for simple function types to be used as event
// handlers.
type HandlerFunc func(*Event)

// HandleEvent calls f.
func (f HandlerFunc) HandleEvent(e *Event) {
	f(e)
}

// MultiHandler is an event handler that fans out events it receives to a list
// of handlers.
type MultiHandler []Handler

// HandleEvent broadcasts e to all handlers of m.
func (m MultiHandler) HandleEvent(e *Event) {
	for _, h := range m {
		h.HandleEvent(e)
	}
}

// Discard is a handler that does nothing with the events it receives.
var Discard Handler = HandlerFunc(func(e *Event) {})
