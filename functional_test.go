package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		prefix string
		strip  bool
	}{
		{"default", "testdata/intput.txt", "testdata/output.txt", "", false},
		{"stripped", "testdata/intput.txt", "testdata/output_strip.txt", "", true},
		{"prefix", "testdata/intput_prefix.txt", "testdata/output_prefix.txt", "prefix ", false},
		{"prefix_strip", "testdata/intput_prefix.txt", "testdata/output_prefix_strip.txt", "prefix ", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			run(in, out, tt.prefix, tt.strip)
			if want, got := string(eb), out.String(); want != got {
				t.Errorf("invalid output:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
