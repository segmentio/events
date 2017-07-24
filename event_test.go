package events

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	t.Run("Clone", func(t *testing.T) {
		e1 := &Event{
			Message: "Hello World",
			Source:  "file.go:42",
			Args:    Args{{"hello", "world"}},
			Time:    time.Now(),
		}
		e2 := e1.Clone()

		if e1 == e2 {
			t.Error("Clone cannot return a value with the same address as the original")
		}

		if !reflect.DeepEqual(e1, e2) {
			t.Errorf("%#v", e2)
		}
	})
}

func TestArgs(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		args := Args{{"hello", "world"}, {"answer", 42}}

		if v, ok := args.Get("answer"); !ok {
			t.Error("expected answer but got nothing")
		} else if !reflect.DeepEqual(v, 42) {
			t.Error("expected answer=42 but got", v)
		}

		if v, ok := args.Get("question"); ok {
			t.Error("expected no question but got", v)
		}
	})
	t.Run("Map", func(t *testing.T) {
		a1 := Args{{"hello", "world"}, {"answer", 42}}
		SortArgs(a1)

		m1 := a1.Map()
		a2 := A(m1)
		SortArgs(a2)

		if !reflect.DeepEqual(a1, a2) {
			t.Error("%#v != %#v", a1, a2)
		}
	})
}

// This test is crafted to crash the program if some of the unsafe operations
// perform illegal memory changes that mess up the GC state.
//
// Run it with CGO_ENABLED=0 GODEBUG=gccheckmark=1 GOTRACEBACK=crash GOGC=1
func TestUnsafe(t *testing.T) {
	logger := Logger{
		Handler:     Discard,
		EnableDebug: true,
	}

	wg := sync.WaitGroup{}

	for i := 0; i != 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i != 2000; i++ {
				_ = make([]byte, 1024*1024)
				var from = "Luke"
				var to = "Han"
				logger.Log("hello world: from=%{from}s, to=%{to}s", from, to)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()
}
