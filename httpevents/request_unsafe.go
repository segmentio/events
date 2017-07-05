// +build go1.5

package httpevents

import (
	"reflect"
	"unsafe"
)

func bytesToStringNonEmpty(b []byte) string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&b[0])),
		Len:  len(b),
	}))
}

func convS2E(s *string) (v interface{}) {
	e := (*eface)(unsafe.Pointer(&v))
	e.t = stringType
	e.v = unsafe.Pointer(s)
	return
}

func convI2E(i *int) (v interface{}) {
	e := (*eface)(unsafe.Pointer(&v))
	e.t = intType
	e.v = unsafe.Pointer(i)
	return
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
)
