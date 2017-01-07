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
		want   bool
	}{
		{"", "", false},
		{"", "2017/01/06 16:25:18 panic: runtime error: invalid memory address or nil pointer dereference", true},
		{"", "2017/01/06 test", true},
		{"", "16:26:44 test", true},
		{"", "16:26:4a test", false},
		{"", "16/26/44 test", false},
		{"", "16:26:44.885183 test", true},
		{"", "16:26:44.88518 test", false},
		{"", "2017/01/06 16:26:44 test", true},
		{"", "2017-01-06 16:26:44 test", false},
		{"", "2017/01/06 16:26:44.885183 test", true},
		{"prefix", "", false},
		{"prefix", "prefix", false},
		{"prefix", "prefix2017/01/06 test", true},
		{"prefix", "prefix16:26:44 test", true},
		{"prefix", "prefix2017/01/06 16:26:44 test", true},
		{"prefix", "prefix2017/01/06 16:26:44.885183 test", true},
		{"prefix", "2017/01/06 16:26:44 test", false},
	}
	for _, tt := range tests {
		if got := IsLog([]byte(tt.line), tt.prefix); got != tt.want {
			t.Errorf("match failed with %q: got %v want %v", tt.line, got, tt.want)
		}
	}
}
