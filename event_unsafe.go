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
	// The conversion of a byte slice ot a string is ensured not to cause a
	// dynamic memory allocation.
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&b[0])),
		Len:  len(b),
	}))
}

func noescape(a []interface{}) (b []interface{}) {
	// The use of reflect.SliceHeader tricks the compiler into thinking that the
	// content of the input slice doesn't escape, so we can prevent dynamic
	// memory allocations that would otherwise happen for each argument of a log
	// message.
	*(*reflect.SliceHeader)(unsafe.Pointer(&b)) = *(*reflect.SliceHeader)(unsafe.Pointer(&a))
	return
}
