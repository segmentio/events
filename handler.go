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

// MultiHandler returns a new Handler which broadcasts the events it receives
// to its list of handlers.
func MultiHandler(handlers ...Handler) Handler {
	c := make([]Handler, len(handlers))
	copy(c, handlers)
	return &multiHandler{
		handlers: c,
	}
}

type multiHandler struct {
	handlers []Handler
}

func (m *multiHandler) HandleEvent(e *Event) {
	for _, h := range m.handlers {
		h.HandleEvent(e)
	}
}

var (
	// Discard is a handler that does nothing with the events it receives.
	Discard Handler = HandlerFunc(func(e *Event) {})

	// DefaultHandler is the default handler used when non is specified.
	DefaultHandler Handler = Discard
)
