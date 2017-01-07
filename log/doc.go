// Package log provides the implementation of a shim between the standard log
// package and event handlers.
//
// The package exposes an API that mimics the standard library, in most cases it
// can be used as a drop in replacement by simply rewriting the import from
// "log" to "github.com/segmentio/events/log".
//
// Importing this package has the side effect of configuring the default output
// of the log package to route events to the default logger of the events
// package.
package log
