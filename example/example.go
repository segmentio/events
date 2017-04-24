package main

import "net/http"
import "github.com/justinas/alice"
import "github.com/segmentio/events"
import "github.com/segmentio/events/httpevents"

func main() {
	http.Handle("/", alice.New(HttpLoggingHandlerConstructor(events.DefaultLogger)).ThenFunc(OK))

	panic(http.ListenAndServe(":8090", nil))
}

func HttpLoggingHandlerConstructor(client *events.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return httpevents.NewHandler(client, h)
	}
}

func OK(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
}
