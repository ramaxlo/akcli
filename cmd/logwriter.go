package cmd

import (
	"bytes"
	"io"
	"sync"
	"time"
)

type timestampWriter struct {
	w   io.Writer
	mu  sync.Mutex
	buf []byte
}

func (tw *timestampWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.buf = append(tw.buf, p...)
	for {
		idx := bytes.IndexByte(tw.buf, '\n')
		if idx < 0 {
			break
		}
		line := tw.buf[:idx+1]
		ts := time.Now().Format("[15:04:05.000] ")
		if _, err := tw.w.Write([]byte(ts)); err != nil {
			return len(p), err
		}
		if _, err := tw.w.Write(line); err != nil {
			return len(p), err
		}
		tw.buf = tw.buf[idx+1:]
	}
	return len(p), nil
}

func (tw *timestampWriter) Flush() error {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if len(tw.buf) > 0 {
		ts := time.Now().Format("[15:04:05.000] ")
		if _, err := tw.w.Write([]byte(ts)); err != nil {
			return err
		}
		if _, err := tw.w.Write(tw.buf); err != nil {
			return err
		}
		if tw.buf[len(tw.buf)-1] != '\n' {
			tw.w.Write([]byte("\n"))
		}
		tw.buf = nil
	}
	return nil
}
