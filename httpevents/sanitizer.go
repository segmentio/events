package httpevents

import "net/http"

type HeaderSanitizer func(http.Header) http.Header
type PathSanitizer func(string) string
type QuerySanitizer func(string) string

type LogSanitizer struct {
	ReqHeaders HeaderSanitizer
	ResHeaders HeaderSanitizer
	Path       PathSanitizer
	Query      QuerySanitizer
}

func NewLogSanitizer() LogSanitizer {
	return LogSanitizer{
		ReqHeaders: DefaultHeaderSanitizer,
		ResHeaders: DefaultHeaderSanitizer,
		Path:       DefaultPathSanitizer,
		Query:      DefaultQuerySanitizer,
	}
}

var DefaultLogSanitizer = NewLogSanitizer()

func (l LogSanitizer) WithQuerySanitizer(sanitizer QuerySanitizer) LogSanitizer {
	l.Query = sanitizer
	return l
}

func (l LogSanitizer) WithReqHeaderSanitizer(sanitizer HeaderSanitizer) LogSanitizer {
	l.ReqHeaders = sanitizer
	return l
}

func (l LogSanitizer) WithResHeaderSanitizer(sanitizer HeaderSanitizer) LogSanitizer {
	l.ResHeaders = sanitizer
	return l
}

func (l LogSanitizer) WithPathSanitizer(sanitizer PathSanitizer) LogSanitizer {
	l.Path = sanitizer
	return l
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
