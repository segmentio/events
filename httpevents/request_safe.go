// +build !go1.5

package httpevents

func bytesToStringNonEmpty(b []byte) string {
	return string(b)
}

func convS2E(s *string) interface{} {
	return *s
}

func convI2E(i *int) interface{} {
	return *i
}
