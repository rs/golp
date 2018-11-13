// Go programs sometime generate output you can't easily control like panics and
// net/http recovered panics. By default, those output contains multiple lines
// with stack traces. This does not play well with most logging systems that will
// generate one log event per outputed line.
//
// The golp is a simple program that reads those kinds of log on its standard
// input, and merge all lines of a given panic or standard multi-lines Go log message
// into a single quoted line.
//
// Usage
//
// Send panics and other program panics to syslog:
//
//     mygoprogram 2>&1 | golp | logger -t mygoprogram -p local7.err
//
// Options:
//
//    -allow-json
//        Allow JSON input not to be escaped. When enabled, max-len is not efforced on JSON lines.
//    -ctx value
//        A key=value to add to the JSON output (can be repeated).
//    -json
//        Wrap messages to one JSON object per line.
//    -json-key string
//        The key name to use for the message in JSON mode. (default "message")
//    -max-len int
//        Strip messages to not exceed this length.
//    -output string
//        A file to append events to. Default output is stdout.
//    -prefix string
//        Go logger prefix set in the application if any.
//    -strip
//        Strip log line timestamps on output.// Send panics and other program panics to syslog:
//
//     mygoprogram 2>&1 | golp | logger -t mygoprogram -p local7.err
//
//     > Jan  8 16:59:26 host mygoprogram: panic: panic: test\n\ngoroutine 1 [running]:\npanic(0x…
//
// Send panics as JSON:
//
//     mygoprogram 2>&1 | golp --json | logger -t mygoprogram -p local7.err
//
//     > Jan  8 16:59:26 host mygoprogram: {"message": "panic: panic: test\n\ngoroutine 1 [running]:\npanic(0x…
// Add context:
//
//     mygoprogram 2>&1 | golp --json --ctx level=error --ctx program=mygoprogram
//
//     > {"level":"error","program":"mygoprogram","message":"panic: panic: test\n\ngoroutine 1 [running]:\npanic(0x…
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/golp/file"
	"github.com/rs/golp/golp"
)

type context map[string]string

func (c *context) String() string {
	return fmt.Sprint(*c)
}

func (c *context) Set(value string) error {
	i := strings.IndexByte(value, '=')
	if i == -1 {
		return errors.New("missing context value")
	}
	(*c)[value[:i]] = value[i+1:]
	return nil
}

func main() {
	maxLen := flag.Int("max-len", 0, "Strip messages to not exceed this length.")
	prefix := flag.String("prefix", "", "Go logger prefix set in the application if any.")
	strip := flag.Bool("strip", false, "Strip log line timestamps on output.")
	json := flag.Bool("json", false, "Wrap messages to one JSON object per line.")
	allowJSON := flag.Bool("allow-json", false, "Allow JSON input not to be escaped. When enabled, max-len is not efforced on JSON lines.")
	jsonKey := flag.String("json-key", "message", "The key name to use for the message in JSON mode.")
	addTimestamp := flag.Bool("add-timestamp", false, "Add a timestamp key to the JSON output (requires json option).")
	output := flag.String("output", "", "A file to append events to. Default output is stdout. "+
		"Use unix: or unixgram: prefix for output on a UNIX socket.")
	ctx := context{}
	flag.Var(&ctx, "ctx", "A key=value to add to the JSON output (can be repeated).")
	flag.Parse()
	if !*json {
		*jsonKey = ""
	}
	var out io.Writer = os.Stdout
	if *output != "" {
		out = file.Output{*output}
	}
	g := golp.Golp{
		In:           os.Stdin,
		Out:          out,
		Context:      ctx,
		MaxLen:       *maxLen,
		Prefix:       *prefix,
		Strip:        *strip,
		AllowJSON:    *allowJSON,
		MessageKey:   *jsonKey,
		AddTimestamp: *addTimestamp,
	}
	g.Run()
}
