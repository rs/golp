// Package event handles incremental building of a log event.
package event

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
)

// Event holds a buffer of a log event content.
type Event struct {
	out   io.Writer
	buf   *bytes.Buffer
	wbuf  []byte
	json  bool
	eol   []byte
	start chan (<-chan time.Time) // timer
	stop  chan bool
	close chan bool
}

var autoFlushCalledHook = func() {}

// New creates an event buffer writing to the out writer on flush.
// When flush, the eol string is appended to the event content.
func New(out io.Writer, eol string, json bool) *Event {
	e := &Event{
		out:   out,
		buf:   bytes.NewBuffer(make([]byte, 0, 4096)),
		wbuf:  make([]byte, 0, 2),
		json:  json,
		eol:   []byte(eol),
		start: make(chan (<-chan time.Time)),
		stop:  make(chan bool),
		close: make(chan bool, 1),
	}
	go e.autoFlushLoop()
	return e
}

// Empty returns true if the event's buffer is empty.
func (e *Event) Empty() bool {
	return e.buf.Len() == 0
}

// Write appends the contents of p to the buffer. The return value
// n is the length of p; err is always nil.
func (e *Event) Write(p []byte) (n int, err error) {
	e.buf.Grow(len(p))
	for _, b := range p {
		e.wbuf = e.wbuf[:0]
		switch b {
		case '"':
			e.wbuf = append(e.wbuf, '\\', b)
		case '\\':
			e.wbuf = append(e.wbuf, `\\`...)
		case '\b':
			e.wbuf = append(e.wbuf, `\b`...)
		case '\f':
			e.wbuf = append(e.wbuf, `\f`...)
		case '\n':
			e.wbuf = append(e.wbuf, `\n`...)
		case '\r':
			e.wbuf = append(e.wbuf, `\r`...)
		case '\t':
			e.wbuf = append(e.wbuf, `\t`...)
		default:
			e.wbuf = append(e.wbuf, b)
		}
		var _n int
		_n, err = e.buf.Write(e.wbuf)
		n += _n
		if err != nil {
			break
		}
	}
	return
}

// Flush appends the eol string to the buffer and copies it to the
// output writer. The buffer is reset after this operation so the
// event can be reused.
//
// If an AutoFlush was in progress, it is stopped by this operation.
func (e *Event) Flush() {
	if e.buf.Len() == 0 {
		return
	}
	e.Stop()
	e.flush()
}

func (e *Event) flush() {
	if e.buf.Len() == 0 {
		return
	}
	if e.json {
		if _, err := fmt.Fprintf(e.out, `{"message": "`); err != nil {
			log.Fatal(err)
		}
	} else {
		e.buf.Write(e.eol)
	}
	if _, err := io.Copy(e.out, e.buf); err != nil {
		log.Fatal(err)
	}
	if e.json {
		if _, err := fmt.Fprintf(e.out, `"}%s`, e.eol); err != nil {
			log.Fatal(err)
		}
	}
	e.buf.Reset()
}

// AutoFlush schedule a flush after delay.
func (e *Event) AutoFlush(delay time.Duration) {
	e.start <- time.After(delay)
}

// Stop clears the auto flush timer
func (e *Event) Stop() {
	e.stop <- true
}

// Close stops the flush loop and release release resources.
func (e *Event) Close() error {
	e.close <- true
	return nil
}

func (e *Event) autoFlushLoop() {
	paused := make(<-chan time.Time) // will never fire
	next := paused
	for {
		select {
		case <-next:
			e.flush()
			autoFlushCalledHook()
		case <-e.stop:
			next = paused
		case timer := <-e.start:
			next = timer
		case <-e.close:
			return
		}
	}
}
