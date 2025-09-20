package sslkeylogfile

import (
	"os"
)

type syncWriter struct {
	f *os.File
}

func newSyncWriter(f *os.File) *syncWriter {
	return &syncWriter{f: f}
}

func (sw *syncWriter) Write(p []byte) (int, error) {
	n, err := sw.f.Write(p)
	if err != nil {
		return n, err
	}
	if err := sw.f.Sync(); err != nil {
		return 0, err
	}
	return n, nil
}
