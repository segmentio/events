// Package ecslogs provides the implementation of an event handler that outputs
// events in a ecslogs-compatible format.
//
// Importing this package has the side effect of configuring the default logger
// to use an ecslogs handler if stdout is not a terminal.
package ecslogs
