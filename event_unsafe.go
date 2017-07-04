// +build go1.5

package events

import (
	"reflect"
	"unsafe"
)

func cloneValue(v interface{}) interface{} {
	// Some sub-packages use optimizations that hack the type system by sharing
	// pointers to avoid generated calls to runtime.convT2E, this functions
	// basically does the opposite and ensures that runtime.convT2E is called
	// to generate new values to ensure no more pointers are shared by cloned
	// events.
	switch x := v.(type) {
	case int:
		return x
	case string:
		return x
	}
	return v
}

func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&b[0])),
		Len:  len(b),
	}))
}
