package golp

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rs/golp/event"
)

func TestRun(t *testing.T) {
	event.TimestampFunc = func() time.Time {
		return time.Time{}
	}
	defer func() {
		event.TimestampFunc = time.Now
	}()
	tests := map[string]struct {
		input        string
		output       string
		maxLen       int
		prefix       string
		strip        bool
		allowJSON    bool
		jsonKey      string
		ctx          map[string]string
		addTimestamp bool
	}{
		"default":        {"testdata/input.txt", "testdata/output.txt", 0, "", false, false, "", nil, false},
		"stripped":       {"testdata/input.txt", "testdata/output_strip.txt", 0, "", true, false, "", nil, false},
		"maxlen":         {"testdata/input.txt", "testdata/output_maxlen.txt", 15, "", true, false, "", nil, false},
		"json_strip":     {"testdata/input.txt", "testdata/output_strip.json", 0, "", true, false, "message", nil, false},
		"json_maxlen":    {"testdata/input.txt", "testdata/output_maxlen.json", 26, "", true, false, "message", nil, false},
		"json_context":   {"testdata/input.txt", "testdata/output_context.json", 0, "", true, false, "message", map[string]string{"foo": "bar"}, false},
		"json_timestamp": {"testdata/input.txt", "testdata/output_timestamp.json", 0, "", true, false, "message", map[string]string{"foo": "bar"}, true},
		"prefix":         {"testdata/input_prefix.txt", "testdata/output_prefix.txt", 0, "prefix ", false, false, "", nil, false},
		"prefix_strip":   {"testdata/input_prefix.txt", "testdata/output_prefix_strip.txt", 0, "prefix ", true, false, "", nil, false},
		"mixed_strip":    {"testdata/input_mixed.txt", "testdata/output_mixed_strip.json", 0, "", true, true, "message", nil, false},
		"mixed_nojson":   {"testdata/input_mixed.txt", "testdata/output_mixed_nojson.json", 0, "", true, false, "message", nil, false},
		"mixed_context":  {"testdata/input_mixed.txt", "testdata/output_mixed_context.json", 0, "", true, true, "message", map[string]string{"foo": "bar"}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			in, err := os.Open(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			expect, err := os.Open(tt.output)
			if err != nil {
				t.Fatal(err)
			}
			defer expect.Close()
			eb, _ := ioutil.ReadAll(expect)
			out := &bytes.Buffer{}
			g := Golp{
				In:           in,
				Out:          out,
				Context:      tt.ctx,
				MaxLen:       tt.maxLen,
				Prefix:       tt.prefix,
				Strip:        tt.strip,
				AllowJSON:    tt.allowJSON,
				MessageKey:   tt.jsonKey,
				AddTimestamp: tt.addTimestamp,
			}
			g.Run()
			if got, want := out.String(), string(eb); want != got {
				t.Errorf("invalid output:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
