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
// 		-prefix string
//         	Go logger prefix set in the application if any.
// 		-strip
//         	Strip log line timestamps on output.
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
	flag.Parse()
	run(os.Stdin, os.Stdout, *prefix, *strip, *json)
}

func run(in io.Reader, out io.Writer, prefix string, strip, json bool) {
	r := bufio.NewReader(in)
	cont := false
	e := event.New(out, "\n", json)
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
