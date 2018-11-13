package file

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// Output is an io.Writer that append each Write into a file at Path. On each
// write the file is open/sync/closed to protect against file rename
// (i.e.: rotation).
type Output struct {
	Path string
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func (o Output) path() (typ, path string) {
	if o.Path == "" || o.Path == "-" {
		return "stdout", ""
	} else if strings.HasPrefix(o.Path, "unix:") {
		return "unix", o.Path[len("unix:"):]
	} else if strings.HasPrefix(o.Path, "unixgram:") {
		return "unixgram", o.Path[len("unixgram:"):]
	} else {
		return "file", o.Path
	}
}

func (o Output) open() (io.WriteCloser, error) {
	typ, path := o.path()
	switch typ {
	case "stdout":
		return nopCloser{os.Stdout}, nil
	case "unix", "unixgram":
		return net.DialUnix(typ, nil, &net.UnixAddr{Net: typ, Name: path})
	case "file":
		return os.OpenFile(o.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	}
	return nil, fmt.Errorf("invalid output")
}

func (o Output) Write(b []byte) (n int, err error) {
	w, err := o.open()
	if err != nil {
		return n, err
	}
	defer w.Close()
	return w.Write(b)
}
