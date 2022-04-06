//go:build go1.5
// +build go1.5

package httpevents

import (
	"unsafe"

	"github.com/segmentio/events/v2"
)

func bytesToStringNonEmpty(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func convS2E(s *string) interface{} {
	return convT2E(stringType, unsafe.Pointer(s))
}

func convI2E(i *int) interface{} {
	return convT2E(intType, unsafe.Pointer(i))
}

func convA2E(a *events.Args) interface{} {
	return convT2E(argsType, unsafe.Pointer(a))
}

func convT2E(t unsafe.Pointer, v unsafe.Pointer) interface{} {
	return *(*interface{})(unsafe.Pointer(&eface{
		t: t,
		v: v,
	}))
}

type eface struct {
	t unsafe.Pointer
	v unsafe.Pointer
}

func typeOf(v interface{}) unsafe.Pointer {
	return (*eface)(unsafe.Pointer(&v)).t
}

var (
	intType    = typeOf(0)
	stringType = typeOf("")
	argsType   = typeOf((events.Args)(nil))
)
