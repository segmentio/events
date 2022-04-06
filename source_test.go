package events

import (
	"runtime"
	"strings"
	"testing"
)

func TestSourceForPC(t *testing.T) {
	pc := [1]uintptr{}
	runtime.Callers(1, pc[:])

	file, line := SourceForPC(pc[0])

	if !strings.HasSuffix(file, "events/source_test.go") {
		t.Error("bad file:", file)
	}

	if line != 11 {
		t.Error("bad line:", line)
	}
}
