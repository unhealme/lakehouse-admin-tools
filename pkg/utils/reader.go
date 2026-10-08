package utils

import (
	"bufio"
	"bytes"
	"io"
	"iter"
)

func ScanLinesSep(sep string) bufio.SplitFunc {
	s := []byte(sep)
	w := len(s)
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, s); i >= 0 {
			return i + w, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}

func IterLinesSeq(r io.Reader, sep string, e chan<- error) iter.Seq[string] {
	return func(yield func(string) bool) {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		if sep != "" {
			sc.Split(ScanLinesSep(sep))
		}
		for sc.Scan() {
			if !yield(sc.Text()) {
				return
			}
		}
		if e != nil {
			if err := sc.Err(); err != nil {
				e <- err
			}
			close(e)
		}
	}
}
