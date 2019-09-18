// +build !go1.5

package httpevents

import "github.com/segmentio/events/v2"

func bytesToStringNonEmpty(b []byte) string {
	return string(b)
}

func convS2E(s *string) interface{} {
	return *s
}

func convI2E(i *int) interface{} {
	return *i
}

func convA2E(a *events.Args) interface{} {
	return *a
}
