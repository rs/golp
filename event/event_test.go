package event

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"
)

func TestFlushEol(t *testing.T) {
	out := &bytes.Buffer{}
	e := New(out, "\n")
	defer e.Close()
	e.Write([]byte("line1"))
	e.Write([]byte("line2"))
	e.Flush()
	if got, want := out.String(), "line1line2\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFlushEmpty(t *testing.T) {
	out := &bytes.Buffer{}
	e := New(out, "\n")
	defer e.Close()
	e.Flush()
	if got, want := out.String(), ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmpty(t *testing.T) {
	e := New(ioutil.Discard, "\n")
	defer e.Close()
	if got, want := e.Empty(), true; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	e.Write([]byte{' '})
	if got, want := e.Empty(), false; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	e.Flush()
	if got, want := e.Empty(), true; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAutoFlush(t *testing.T) {
	done := make(chan bool, 1)
	autoFlushCalledHook = func() {
		done <- true
	}
	defer func() {
		autoFlushCalledHook = func() {}
	}()
	out := &bytes.Buffer{}
	e := New(out, "\n")
	defer e.Close()
	e.Write([]byte{'x'})
	c := make(chan time.Time)
	e.start <- c // simulate AutoFlush()
	if got, want := out.String(), ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	go func() {
		c <- time.Time{}
	}()
	<-done
	if got, want := out.String(), "x\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
