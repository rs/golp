package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	tests := map[string]struct {
		input  string
		output string
		prefix string
		strip  bool
		json   bool
	}{
		"default":      {"testdata/intput.txt", "testdata/output.txt", "", false, false},
		"stripped":     {"testdata/intput.txt", "testdata/output_strip.txt", "", true, false},
		"json_strip":   {"testdata/intput.txt", "testdata/output_strip.json", "", true, true},
		"prefix":       {"testdata/intput_prefix.txt", "testdata/output_prefix.txt", "prefix ", false, false},
		"prefix_strip": {"testdata/intput_prefix.txt", "testdata/output_prefix_strip.txt", "prefix ", true, false},
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
			run(in, out, tt.prefix, tt.strip, tt.json)
			if want, got := string(eb), out.String(); want != got {
				t.Errorf("invalid output:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
