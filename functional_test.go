package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	tests := map[string]struct {
		input   string
		output  string
		maxLen  int
		prefix  string
		strip   bool
		jsonKey string
		ctx     map[string]string
	}{
		"default":      {"testdata/intput.txt", "testdata/output.txt", 0, "", false, "", nil},
		"stripped":     {"testdata/intput.txt", "testdata/output_strip.txt", 0, "", true, "", nil},
		"maxlen":       {"testdata/intput.txt", "testdata/output_maxlen.txt", 15, "", true, "", nil},
		"json_strip":   {"testdata/intput.txt", "testdata/output_strip.json", 0, "", true, "message", nil},
		"json_maxlen":  {"testdata/intput.txt", "testdata/output_maxlen.json", 26, "", true, "message", nil},
		"json_context": {"testdata/intput.txt", "testdata/output_context.json", 0, "", true, "message", map[string]string{"foo": "bar"}},
		"prefix":       {"testdata/intput_prefix.txt", "testdata/output_prefix.txt", 0, "prefix ", false, "", nil},
		"prefix_strip": {"testdata/intput_prefix.txt", "testdata/output_prefix_strip.txt", 0, "prefix ", true, "", nil},
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
			run(in, out, tt.ctx, tt.maxLen, tt.prefix, tt.strip, tt.jsonKey)
			if want, got := string(eb), out.String(); want != got {
				t.Errorf("invalid output:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
