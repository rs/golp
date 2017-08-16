package golp

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/rs/golp/event"
	"github.com/rs/golp/parser"
)

type Golp struct {
	In         io.Reader
	Out        io.Writer
	Context    map[string]string
	MaxLen     int
	Prefix     string
	Strip      bool
	AllowJSON  bool
	MessageKey string
}

func (g Golp) Run() {
	r := bufio.NewReader(g.In)
	cont := false
	e, err := event.New(g.Out, g.Context, g.MaxLen, "\n", g.MessageKey)
	if err != nil {
		log.Fatal(err)
	}
	e.AllowJSON = g.AllowJSON
	autoFlushDelay := 5 * time.Millisecond
	go func() {
		// Flush before exit
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		e.Flush()
		os.Exit(1)
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
			} else if index := parser.IsLog(line, g.Prefix); index > 0 {
				// Flush previous event if any
				e.Flush()
				if g.Strip {
					// Strip log message header (prefix, timestamp)
					line = line[index:]
				}
			} else if g.AllowJSON && parser.IsJSON(line) {
				// Flush previous event if any
				e.Flush()
				e.Write(line)
				e.Flush()
				continue
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
