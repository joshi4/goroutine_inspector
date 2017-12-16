package goroutine_inspector

import (
	"fmt"
	"testing"
)

func TestGoroutineLeaks(t *testing.T) {
	tr, err := Start()
	if err != nil {
		t.Error(err)
	}

	ch := make(chan bool)
	go routine(ch)
	<-ch

	// leak three go routines
	go routine(make(chan bool))
	go routine(make(chan bool))
	go routine(make(chan bool))

	leaks, err := tr.GoroutineLeaks()
	if err != nil {
		t.Error(err)
	}

	if len(leaks) != 3 {
		t.Errorf("goroutine_leaks = %d, want = %d", len(leaks), 3)
	}
	fmt.Println(leaks)
}

func routine(ch chan bool) {
	ch <- false
}