package golp

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	tests := map[string]struct {
		input     string
		output    string
		maxLen    int
		prefix    string
		strip     bool
		allowJSON bool
		jsonKey   string
		ctx       map[string]string
	}{
		"default":       {"testdata/input.txt", "testdata/output.txt", 0, "", false, false, "", nil},
		"stripped":      {"testdata/input.txt", "testdata/output_strip.txt", 0, "", true, false, "", nil},
		"maxlen":        {"testdata/input.txt", "testdata/output_maxlen.txt", 15, "", true, false, "", nil},
		"json_strip":    {"testdata/input.txt", "testdata/output_strip.json", 0, "", true, false, "message", nil},
		"json_maxlen":   {"testdata/input.txt", "testdata/output_maxlen.json", 26, "", true, false, "message", nil},
		"json_context":  {"testdata/input.txt", "testdata/output_context.json", 0, "", true, false, "message", map[string]string{"foo": "bar"}},
		"prefix":        {"testdata/input_prefix.txt", "testdata/output_prefix.txt", 0, "prefix ", false, false, "", nil},
		"prefix_strip":  {"testdata/input_prefix.txt", "testdata/output_prefix_strip.txt", 0, "prefix ", true, false, "", nil},
		"mixed_strip":   {"testdata/input_mixed.txt", "testdata/output_mixed_strip.json", 0, "", true, true, "message", nil},
		"mixed_nojson":  {"testdata/input_mixed.txt", "testdata/output_mixed_nojson.json", 0, "", true, false, "message", nil},
		"mixed_context": {"testdata/input_mixed.txt", "testdata/output_mixed_context.json", 0, "", true, true, "message", map[string]string{"foo": "bar"}},
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
				In:         in,
				Out:        out,
				Context:    tt.ctx,
				MaxLen:     tt.maxLen,
				Prefix:     tt.prefix,
				Strip:      tt.strip,
				AllowJSON:  tt.allowJSON,
				MessageKey: tt.jsonKey,
			}
			g.Run()
			if got, want := out.String(), string(eb); want != got {
				t.Errorf("invalid output:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
