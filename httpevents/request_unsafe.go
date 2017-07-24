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

func convS2E(s *string) interface{} {
	return *(*interface{})(unsafe.Pointer(&eface{
		t: stringType,
		v: unsafe.Pointer(s),
	}))
}

func convI2E(i *int) interface{} {
	return *(*interface{})(unsafe.Pointer(&eface{
		t: intType,
		v: unsafe.Pointer(i),
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
)
