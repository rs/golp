// Package parser provides some utility functions to recognise panic and log lines.
package parser

import "bytes"

var (
	panicPrefix       = []byte("panic: ")
	logPrefixPatterns = [][]byte{
		[]byte("2000/01/02 "),
		[]byte("12:00:00 "),
		[]byte("12:00:00.000000 "),
		[]byte("2000/01/02 12:00:00 "),
		[]byte("2000/01/02 12:00:00.000000 "),
	}
)

// IsPanic returns true if the line is the first line of a Go panic
func IsPanic(line []byte) bool {
	return bytes.HasPrefix(line, panicPrefix)
}

// IsLog returns true if the line is the first line of log produced
// by the Go logger
func IsLog(line []byte, prefix string) bool {
	// example: 2017/01/06 14:16:13 log line
	if len(line) < len(prefix) {
		return false
	}
	line = line[len(prefix):]
	for _, pattern := range logPrefixPatterns {
		if matchPattern(line, pattern) {
			return true
		}
	}
	return false
}

// matchPattern return true if the given line starts with the given pattern.
// The pattern match if all non numeric characters match and if all numeric
// character are (non necessary equal) numbers on both sides.
func matchPattern(line []byte, pattern []byte) bool {
	if len(line) < len(pattern) {
		return false
	}
	for i, b := range pattern {
		if isNumber(b) {
			if !isNumber(line[i]) {
				return false
			}
		} else {
			if b != line[i] {
				return false
			}
		}
	}
	return true
}

// isNumber returns true if b is an ASCII char between 0 and 9.
func isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}
