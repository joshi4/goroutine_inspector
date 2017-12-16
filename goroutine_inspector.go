package goroutine_inspector

import (
	"bytes"
	"regexp"
	"runtime/trace"
	"strconv"
	"sync"

	. "github.com/joshi4/goroutine_inspector/internal/trace"
)

type Trace struct {
	// start synOnce
	buf *bytes.Buffer

	done   sync.Once
	donech chan struct{}
	err    error
}

// Start starts a trace for inspection.
//
// NOTE: Start must only be called once per executable
func Start() (*Trace, error) {
	t := &Trace{
		buf:    new(bytes.Buffer),
		donech: make(chan struct{}),
	}

	if err := trace.Start(t.buf); err != nil {
		return nil, err
	}
	return t, nil
}

func shouldAddEvent(e *Event) bool {
	pattern := "github.com/joshi4/goroutine_inspector.Start|runtime/trace.Start"
	ok, _ := regexp.MatchString(pattern, peekFn(e.Stk))
	return !ok
}

func peekFn(s []*Frame) string {
	if len(s) == 0 {
		return ""
	}
	return s[0].Fn
}

// Stop stops the trace.
func (t *Trace) Stop() {
	t.done.Do(func() {
		trace.Stop()
		close(t.donech)
	})
}

// GoroutineLeaks returns all go routines that were created
// but did not terminate during the trace period.
// GoroutineLeaks calls Stop()
func (t *Trace) GoroutineLeaks() ([]string, error) {
	t.Stop()

	leakedGoRoutines := make([]string, 0)

	events, err := Parse(t.buf, "")
	if err != nil {
		return nil, err
	}

	for _, e := range events {
		if e.Type == EvGoCreate && goroutineLeaked(e) {
			if shouldAddEvent(e) {
				leakedGoRoutines = append(leakedGoRoutines, printStack(e.Stk))
			}
		}
	}
	return leakedGoRoutines, nil
}

func goroutineLeaked(e *Event) bool {
	if e.Link == nil {
		return e.Type != EvGoEnd
	}
	return goroutineLeaked(e.Link)
}

func printStack(s []*Frame) string {
	str := ""
	for _, fr := range s {
		str += "function:" + fr.Fn + "\n" + "file:" + fr.File + "\nline:" + strconv.Itoa(fr.Line) + "\n"
	}
	return str + "end\n"
}