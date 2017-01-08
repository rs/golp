package parser

import "testing"

func TestIsPanic(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"panic: runtime error: invalid memory address or nil pointer dereference", true},
		{"panic: ", true},
		{"panic:", false},
		{"2017/01/06 16:25:18 panic: runtime error: invalid memory address or nil pointer dereference", false},
	}
	for _, tt := range tests {
		if got := IsPanic([]byte(tt.line)); got != tt.want {
			t.Errorf("match failed with %q: got %v want %v", tt.line, got, tt.want)
		}
	}
}

func TestIsLog(t *testing.T) {
	tests := []struct {
		prefix string
		line   string
		want   int
	}{
		{"", "", -1},
		{"", "2017/01/06 16:25:18 panic: runtime error: invalid memory address or nil pointer dereference", 20},
		{"", "2017/01/06 test", 11},
		{"", "16:26:44 test", 9},
		{"", "16:26:4a test", -1},
		{"", "16/26/44 test", -1},
		{"", "16:26:44.885183 test", 16},
		{"", "16:26:44.88518 test", -1},
		{"", "2017/01/06 16:26:44 test", 20},
		{"", "2017-01-06 16:26:44 test", -1},
		{"", "2017/01/06 16:26:44.885183 test", 27},
		{"prefix", "", -1},
		{"prefix", "prefix", -1},
		{"prefix", "prefix2017/01/06 test", 17},
		{"prefix", "prefix16:26:44 test", 15},
		{"prefix", "prefix2017/01/06 16:26:44 test", 26},
		{"prefix", "prefix2017/01/06 16:26:44.885183 test", 33},
		{"prefix", "2017/01/06 16:26:44 test", -1},
	}
	for _, tt := range tests {
		if got := IsLog([]byte(tt.line), tt.prefix); got != tt.want {
			t.Errorf("match failed with %q: got %v want %v", tt.line, got, tt.want)
		}
	}
}
