// Go programs sometime generate output you can't easily control like panics and
// net/http recovered panics. By default, those output contains multiple lines
// with stack traces. This does not play well with most logging systems that will
// generate one log event per outputed line.
//
// The golp is a simple program that reads those kinds of log on its standard
// input, and merge all lines of a given panic or standard multi-lines Go log message
// into a single quotted line.
//
// Usage
//
// Send panics and other program panics to syslog:
//
//     mygoprogram 2>&1 | golp | logger -t mygoprogram -p local7.err
//
// Options:
//
// 		-json
//         	Wrap messages to JSON one object per line.
//  	-json-key string
//  		The key name to use for the message in JSON mode. (default "message")
// 		-prefix string
//         	Go logger prefix set in the application if any.
// 		-strip
//         	Strip log line timestamps on output.
//
// Send panics and other program panics to syslog:
//
//     mygoprogram 2>&1 | golp | logger -t mygoprogram -p local7.err
//
//     > Jan  8 16:59:26 host panic: panic: test\n\ngoroutine 1 [running]:\npanic(0x…
//
// Send panics as JSON:
//
//     mygoprogram 2>&1 | golp --json | logger -t mygoprogram -p local7.err
//
//     > Jan  8 16:59:26 host {"message": "panic: panic: test\n\ngoroutine 1 [running]:\npanic(0x…
//
// Add some fields to the JSON output (using [jq](https://stedolan.github.io/jq/)):
//
//     mygoprogram 2>&1 | golp --json | jq -c '. + {"level": "error", "program": "mygoprogram"}'
//
//     > {"level": "error", "program": "mygoprogram", "message": "panic: panic: test\n\ngoroutine 1 [running]:\npanic(0x…
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/rs/golp/event"
	"github.com/rs/golp/parser"
)

func main() {
	prefix := flag.String("prefix", "", "Go logger prefix set in the application if any.")
	strip := flag.Bool("strip", false, "Strip log line timestamps on output.")
	json := flag.Bool("json", false, "Wrap messages to JSON one object per line.")
	jsonKey := flag.String("json-key", "message", "The key name to use for the message in JSON mode.")
	flag.Parse()
	if !*json {
		*jsonKey = ""
	}
	run(os.Stdin, os.Stdout, *prefix, *strip, *jsonKey)
}

func run(in io.Reader, out io.Writer, prefix string, strip bool, jsonKey string) {
	r := bufio.NewReader(in)
	cont := false
	e := event.New(out, "\n", jsonKey)
	autoFlushDelay := 5 * time.Millisecond
	go func() {
		// Flush before exit
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		e.Flush()
	}()
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			e.Flush()
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		// Stop the previous auto-flush if any so we don't accidently flush
		// before reading the new line.
		e.Stop()
		if !cont {
			if parser.IsPanic(line) {
				// Flush previous event if any
				e.Flush()
			} else if index := parser.IsLog(line, prefix); index > 0 {
				// Flush previous event if any
				e.Flush()
				if strip {
					// Strop log message header (prefix, timestamp)
					line = line[index:]
				}
			} else if !e.Empty() {
				// The line is a continuation, add a quoted carriage return before
				// appending it to the current event.
				e.Write([]byte{'\n'})
			}
		}
		e.Write(line)
		// Auto-flush the event after if no new line is read for the given delay.
		e.AutoFlush(autoFlushDelay)
		cont = isPrefix
	}
}
