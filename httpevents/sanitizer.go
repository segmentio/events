package httpevents

import "net/http"

// HeaderSanitizer is a function that sanitizes a header
type HeaderSanitizer func(http.Header) http.Header

// PathSanitizer is a function that sanitizes a path
type PathSanitizer func(string) string

// QuerySanitizer is a function that sanitizes a query
type QuerySanitizer func(string) string

type LogSanitizer struct {
	reqHeaders HeaderSanitizer
	resHeaders HeaderSanitizer
	Path       PathSanitizer
	Query      QuerySanitizer
}

func NewLogSanitizer() LogSanitizer {
	return LogSanitizer{
		reqHeaders: DefaultHeaderSanitizer,
		resHeaders: DefaultHeaderSanitizer,
		Path:       DefaultPathSanitizer,
		Query:      DefaultQuerySanitizer,
	}
}

var DefaultLogSanitizer = NewLogSanitizer()

// ReqHeaders makes a copy of the request headers and sanitizes the resulting headers
func (l LogSanitizer) ReqHeaders(h http.Header) http.Header {
	// Copy the existing headers so we aren't mutating the original
	return l.reqHeaders(copyMap(h))
}

// ResHeaders makes a copy of the response headers and sanitizes the resulting headers
func (l LogSanitizer) ResHeaders(h http.Header) http.Header {
	// Copy the existing headers so we aren't mutating the original
	return l.resHeaders(copyMap(h))
}

func (l LogSanitizer) WithQuerySanitizer(sanitizer QuerySanitizer) LogSanitizer {
	l.Query = sanitizer
	return l
}

func (l LogSanitizer) WithReqHeaderSanitizer(sanitizer HeaderSanitizer) LogSanitizer {
	l.reqHeaders = sanitizer
	return l
}

func (l LogSanitizer) WithResHeaderSanitizer(sanitizer HeaderSanitizer) LogSanitizer {
	l.resHeaders = sanitizer
	return l
}

func (l LogSanitizer) WithPathSanitizer(sanitizer PathSanitizer) LogSanitizer {
	l.Path = sanitizer
	return l
}

func copyMap(h http.Header) http.Header {
	m := make(http.Header, len(h))
	for k, v := range h {
		m[k] = v
	}
	return m
}

var DefaultHeaderSanitizer HeaderSanitizer = func(h http.Header) http.Header {
	return h
}

var DefaultPathSanitizer PathSanitizer = func(path string) string {
	return path
}

var DefaultQuerySanitizer QuerySanitizer = func(query string) string {
	return query
}
