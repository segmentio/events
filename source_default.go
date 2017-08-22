// +build !go1.7

package events

import "runtime"

func fileLineFunc(pc uintptr) (file string, line int, name string) {
	fn := runtime.FuncForPC(pc)
	file, line = fn.FileLine(pc)
	name = fn.Name()
	return
}
