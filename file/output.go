package file

import "os"

// Output is an io.Writer that append each Write into a file at Path. On each
// write the file is open/sync/closed to protect against file rename
// (i.e.: rotation).
type Output struct {
	Path string
}

func (o Output) Write(b []byte) (n int, err error) {
	var f *os.File
	f, err = os.OpenFile(o.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	return f.Write(b)
}
