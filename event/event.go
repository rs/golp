// Package event handles incremental building of a log event.
package event

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"time"
)

// Event holds a buffer of a log event content.
type Event struct {
	out        *bufio.Writer
	buf        *bytes.Buffer
	wbuf       []byte
	maxLen     int
	exceeded   int
	allowJSON  bool
	prefix     []byte
	suffix     []byte
	isJSON     bool
	jsonPrefix []byte
	jsonSuffix []byte
	timePrefix []byte
	timeFormat string
	write      chan func()
	flush      chan chan bool
	start      chan (<-chan time.Time) // timer
	stop       chan bool
	close      chan bool
}

// TimestampFunc is called to generate timestamps.
var TimestampFunc = time.Now

type Option func(e *Event) error

var autoFlushCalledHook = func() {}

// New creates an event buffer writing to the out writer on flush.
func New(out io.Writer, options ...Option) (e *Event, err error) {
	e = &Event{
		out:        bufio.NewWriterSize(out, 4096),
		buf:        bytes.NewBuffer(make([]byte, 0, 4096)),
		wbuf:       make([]byte, 0, 2),
		maxLen:     0,
		write:      make(chan func()),
		flush:      make(chan chan bool),
		start:      make(chan (<-chan time.Time)),
		stop:       make(chan bool),
		close:      make(chan bool, 1),
		jsonSuffix: []byte("\n"),
		suffix:     []byte("\n"),
	}
	for _, option := range options {
		if err := option(e); err != nil {
			return nil, err
		}
	}
	if e.maxLen > 0 {
		minPayload := len(e.prefix) + len(e.suffix)
		if len(e.timePrefix) > 0 {
			minPayload += len(e.timePrefix) + len(e.timeFormat)
		}
		if e.maxLen < minPayload {
			return nil, errors.New("max len is lower than JSON envelope")
		}
	}
	go e.writeLoop()
	return
}

// AllowJSON allows JSON input. When this option is true and the input is JSON,
// the maxlen option can not be enforced.
func AllowJSON(enabled bool, context map[string]string) Option {
	return func(e *Event) error {
		e.allowJSON = enabled
		if len(context) > 0 {
			ctxJSON, err := json.Marshal(context)
			if err != nil {
				return err
			}
			// Prepare for embedding by removing { } and append a comma
			ctxJSON[len(ctxJSON)-1] = ','
			e.jsonPrefix = ctxJSON // store {ctx, for insertion when input is already JSON
		} else {
			e.jsonPrefix = []byte{'{'}
		}
		return nil
	}
}

// JSONOutput makes the event output formatted as JSON. The content of the
// message is written as the messageKey key and the context is added to the JSON
// object.
func JSONOutput(messageKey string, context map[string]string) Option {
	return func(e *Event) (err error) {
		if messageKey == "" {
			messageKey = "msg"
		}
		var ctxJSON []byte
		if len(context) > 0 {
			ctxJSON, err = json.Marshal(context)
			if err != nil {
				return
			}
			// Prepare for embedding by removing { } and append a comma
			ctxJSON[len(ctxJSON)-1] = ','
			ctxJSON = ctxJSON[1:]
		}
		e.prefix = []byte(fmt.Sprintf(`{%s"%s":"`, ctxJSON, messageKey))
		e.suffix = []byte("\"}\n")
		return
	}
}

// AddTimestamp adds a timestamp to each event using the provided format.
// If the output is json, the value is added to the jsonKey key.
// If JSON input is allowed and input is JSON, no timestamp is added.
// JSONOutput must be used before this option.
func AddTimestamp(jsonKey, format string) Option {
	return func(e *Event) error {
		if len(e.prefix) == 0 {
			return errors.New("AddTimestamp used before JSONOutput")
		}
		e.timePrefix = []byte(fmt.Sprintf(`","%s":`, jsonKey))
		e.timeFormat = format
		e.suffix = []byte("}\n")
		return nil
	}
}

// MaxLen defines a maximum len for the output event. If the event is larger,
// the message is truncated to fix into maxLen.
func MaxLen(maxLen int) Option {
	return func(e *Event) error {
		e.maxLen = maxLen
		return nil
	}
}

// Empty returns true if the event's buffer is empty.
func (e *Event) Empty() bool {
	return e.buf.Len() == 0
}

// Write appends the contents of p to the buffer. The return value
// n is the length of p; err is always nil.
func (e *Event) Write(p []byte) (n int, err error) {
	done := make(chan struct{})
	e.write <- (func() {
		n, err = e.doWrite(p)
		close(done)
	})
	<-done
	return
}

// isJSON returns true if b *seems* to contain a JSON object
func isJSON(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	return b[0] == '{' && b[1] == '"'
}

func (e *Event) doWrite(p []byte) (n int, err error) {
	if e.allowJSON && !e.isJSON && e.buf.Len() == 0 {
		// Check if the line start as a JSON object.
		// If JSON, insert the context and write directly to the output.
		e.isJSON = isJSON(p)
		if e.isJSON {
			e.out.Write(e.jsonPrefix)
			n, err = e.out.Write(p[1:]) // skip the {
			return n + 1, err
		}
	}
	if e.isJSON {
		// Input is already JSON, do not escape or compute exceeding
		return e.out.Write(p)
	}
	if e.exceeded > 0 {
		e.exceeded += len(p)
		return
	}
	overhead := len(e.prefix) + len(e.suffix)
	e.buf.Grow(len(p))
	for i, b := range p {
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
		if e.maxLen > 0 && e.buf.Len()+overhead+len(e.wbuf) > e.maxLen {
			e.exceeded = len(p) - i
			break
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
	if e.buf.Len() == 0 && !e.isJSON {
		return
	}
	c := make(chan bool)
	// Make the flushLoop to flush
	e.flush <- c
	// Wait for the flush to end
	<-c
}

// uintLen return the number of chars taken by an integer
func uintLen(i uint) (l int) {
	if i == 0 {
		return 1
	}
	return int(math.Log10(float64(i))) + 1
}

func (e *Event) doFlush() {
	defer func() {
		if err := e.out.Flush(); err != nil {
			logWriteErr(err)
		}
	}()
	if e.isJSON {
		e.isJSON = false
		if _, err := e.out.Write(e.jsonSuffix); err != nil {
			logWriteErr(err)
		}
		return
	}
	if e.buf.Len() == 0 {
		return
	}
	if len(e.prefix) > 0 {
		if _, err := e.out.Write(e.prefix); err != nil {
			logWriteErr(err)
		}
	}
	const elipse = "[]..."
	if e.exceeded > 0 && e.buf.Len() > len(elipse)+1 {
		// Insert [total_bytes_truncated]… at the end of the message if possible
		msg := e.buf.Bytes()
		// estimate truncated byte number including the marker
		t := e.exceeded + len(elipse)
		t += uintLen(uint(t))
		if pos := len(msg) - (t - e.exceeded); pos > 0 {
			// Ensure we don't cut in the middle of an escaped char by
			// searching for the first \ of a continuous sequence of \
			// and consider removing the current one if is not an escaped
			// char itself
			escapes := 0
			for pos-escapes > 0 && msg[pos-escapes] == '\\' {
				escapes++
			}
			if escapes > 0 {
				pos -= (escapes + 1) % 2
			}
			// Compute the actual truncated bytes before escaping
			t := e.exceeded
			for i := pos; i < len(msg); i++ {
				if msg[i] == '\\' {
					// Skip escaped char from the count
					i++
				}
				t++
			}
			eb := strconv.FormatInt(int64(t), 10)
			msg = append(msg[:pos], elipse[0])
			msg = append(msg, eb...)
			msg = append(msg, elipse[1:]...)
		}
		if _, err := e.out.Write(msg); err != nil {
			logWriteErr(err)
		}
	} else {
		if _, err := io.Copy(e.out, e.buf); err != nil {
			logWriteErr(err)
		}
	}
	if len(e.timePrefix) > 0 {
		if _, err := e.out.Write(e.timePrefix); err != nil {
			logWriteErr(err)
		}
		ts := strconv.Quote(TimestampFunc().Format(e.timeFormat))
		if _, err := e.out.WriteString(ts); err != nil {
			logWriteErr(err)
		}
	}
	if len(e.suffix) > 0 {
		if _, err := e.out.Write(e.suffix); err != nil {
			logWriteErr(err)
		}
	}
	e.buf.Reset()
	e.exceeded = 0
}

func logWriteErr(err error) {
	log.Printf("golp: write error: %v", err)
}

// AutoFlush schedule a flush after delay.
func (e *Event) AutoFlush(delay time.Duration) {
	e.start <- time.After(delay)
}

// Stop clears the auto flush timer
func (e *Event) Stop() {
	e.stop <- true
}

// Close stops the flush loop and releases resources.
func (e *Event) Close() error {
	close(e.close)
	return nil
}

func (e *Event) writeLoop() {
	paused := make(<-chan time.Time) // will never fire
	next := paused
	for {
		select {
		case cmd := <-e.write:
			cmd()
		case done := <-e.flush:
			e.doFlush()
			next = paused // cancel the autoflush
			close(done)   // notify caller
		case <-next:
			e.doFlush()
			next = paused
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
