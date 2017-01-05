package events

import (
	"runtime"
	"testing"
)

func TestSourceForPC(t *testing.T) {
	pc := [1]uintptr{}
	runtime.Callers(1, pc[:])

	file, line := SourceForPC(pc[0])

	if file != "github.com/segmentio/events/source_test.go" {
		t.Error("bad file:", file)
	}

	if line != 12 {
		t.Error("bad line:", line)
	}
}
