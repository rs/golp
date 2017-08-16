package event

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func TestWriteEscape(t *testing.T) {
	e, _ := New(ioutil.Discard)
	defer e.Close()
	e.Write([]byte("\b\f\r\n\t\\\""))
	if got, want := e.buf.String(), `\b\f\r\n\t\\\"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteMaxLen(t *testing.T) {
	e, _ := New(ioutil.Discard, MaxLen(5))
	defer e.Close()
	n, _ := e.Write([]byte("abcdefghij"))
	if got, want := n, 4; got != want {
		t.Errorf("invalid n: got %v, want %v", got, want)
	}
	if got, want := e.buf.String(), "abcd"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteMaxLenEscaping(t *testing.T) {
	tests := []struct {
		maxLen       int
		input        string
		output       string
		len          int
		outputMarker string
		lenMarker    int
	}{
		{1, "abcd", ``, 0, ``, 0},
		{10, "abcdefghijklmnopqrstuvwxyz", `abcdefghi`, 9, "ab[24]…\n", 10},
		{10, "ab\\cdf\n\n\n\n\n", `ab\\cdf\n`, 9, "ab[9]…\n", 9},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			out := &bytes.Buffer{}
			e, _ := New(out, MaxLen(tt.maxLen))
			defer e.Close()
			n, _ := e.Write([]byte(tt.input))
			if got, want := n, tt.len; got != want {
				t.Errorf("invalid n: got %v, want %v", got, want)
			}
			if got, want := e.buf.String(), tt.output; got != want {
				t.Errorf("invalid buffer content: got %q, want %q", got, want)
			}
			e.Flush()
			if got, want := out.Len(), tt.lenMarker; got != want {
				t.Errorf("invalid length with marker: got %v, want %v", got, want)
			}
			if got, want := out.String(), tt.outputMarker; got != want {
				t.Errorf("invalid output with marker: got %q, want %q", got, want)
			}
		})
	}
}

func TestFlushJSON(t *testing.T) {
	out := &bytes.Buffer{}
	e, _ := New(out, JSONOutput("message", nil))
	defer e.Close()
	e.Write([]byte("line1\n"))
	e.Write([]byte("line2"))
	e.Flush()
	if got, want := out.String(), "{\"message\":\"line1\\nline2\"}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFlushJSONMaxLen(t *testing.T) {
	out := &bytes.Buffer{}
	e, _ := New(out, MaxLen(33), JSONOutput("message", nil))
	defer e.Close()
	e.Write([]byte("line1\n"))
	e.Write([]byte("line2\n"))
	e.Write([]byte("line3"))
	e.Flush()
	if got, want := len(out.String()), 33; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := out.String(), "{\"message\":\"line1\\nline2[6]…\"}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFlushEmpty(t *testing.T) {
	out := &bytes.Buffer{}
	e, _ := New(out)
	defer e.Close()
	e.Flush()
	if got, want := out.String(), ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAllowJSON(t *testing.T) {
	out := &bytes.Buffer{}
	e, _ := New(out, AllowJSON(true, nil))
	defer e.Close()
	e.Write([]byte(`{"foo":"bar"}`))
	e.Flush()
	if got, want := out.String(), "{\"foo\":\"bar\"}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	out.Reset()
	e, _ = New(out)
	e.Write([]byte(`{"foo":"bar"}`))
	e.Flush()
	if got, want := out.String(), "{\\\"foo\\\":\\\"bar\\\"}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmpty(t *testing.T) {
	e, _ := New(ioutil.Discard)
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
	e, _ := New(out)
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
