//go:build go1.5
// +build go1.5

package events

import (
	"reflect"
	"unsafe"
)

func cloneValue(v interface{}) interface{} {
	// Some sub-packages use optimizations that hack the type system by sharing
	// pointers to avoid generated calls to runtime.convT2E, this functions
	// basically does the opposite and ensures that a new value is created.
	rv := reflect.ValueOf(v)
	nv := reflect.New(rv.Type()).Elem()
	nv.Set(rv)
	return nv.Interface()
}

func bytesToString(b []byte) string {
	// The conversion of a byte slice to a string is ensured not to cause a
	// dynamic memory allocation.
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&b))
}
