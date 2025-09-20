package sslkeylogfile

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

const (
	envSSLKey  = "SSLKEYLOGFILE"
	envVerbose = "SSLKEYLOGFILE_VERBOSE"
)

var (
	globalWriter *keyLogWriter
)

type keyLogWriter struct {
	pattern string
	counter atomic.Int64
}

func (w *keyLogWriter) NewKeyWriter() (io.Writer, error) {
	cur := w.counter.Add(1)
	filename := w.pattern
	if cur > 1 {
		filename = w.pattern + "." + strconv.FormatInt(cur, 10)
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("sslkeylogfile: failed to create key log file %q: %w", filename, err)
	}
	if verboseEnabled() {
		fmt.Fprintf(os.Stderr, "sslkeylogfile: writing TLS keys to %s\n", filename)
	}
	return newSyncWriter(f), nil
}

// NewTLSConfig returns a tls.Config with KeyLogWriter set to a writer that
// writes to the file specified in the SSLKEYLOGFILE environment variable.
// If SSLKEYLOGFILE is not set, NewTLSConfig returns an empty tls.Config.
func NewTLSConfig() (*tls.Config, error) {
	w, err := newGlobalWriter()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		KeyLogWriter: w,
	}, nil
}

// NewTransport returns an http.Transport with TLSClientConfig.KeyLogWriter
// set to a writer that writes to the file specified in the SSLKEYLOGFILE
// environment variable. If SSLKEYLOGFILE is not set, NewTransport returns a
// default http.Transport.
func NewTransport() (*http.Transport, error) {
	w, err := newGlobalWriter()
	if err != nil {
		return nil, err
	}
	if w == nil {
		return &http.Transport{}, nil
	}
	return &http.Transport{
		TLSClientConfig: &tls.Config{KeyLogWriter: w},
	}, nil
}

// PatchDefaultTransport patches http.DefaultTransport to log SSL/TLS keys to
// the file specified in the SSLKEYLOGFILE environment variable. If
// SSLKEYLOGFILE is not set, PatchDefaultTransport does nothing.
func PatchDefaultTransport() error {
	w, err := newGlobalWriter()
	if err != nil {
		return err
	}
	if w != nil {
		t, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return fmt.Errorf("sslkeylogfile: cannot configure http.DefaultTransport, it's not an *http.Transport (got %T)", http.DefaultTransport)
		}
		if t.TLSClientConfig != nil {
			t.TLSClientConfig = t.TLSClientConfig.Clone()
		} else {
			t.TLSClientConfig = &tls.Config{}
		}
		t.TLSClientConfig.KeyLogWriter = w
		if verboseEnabled() {
			fmt.Fprintf(os.Stderr, "sslkeylogfile: enabled for http.DefaultTransport\n")
		}
	}
	return nil
}

func newGlobalWriter() (io.Writer, error) {
	if globalWriter == nil {
		return nil, nil
	}
	return globalWriter.NewKeyWriter()
}

func verboseEnabled() bool {
	v := os.Getenv(envVerbose)
	enabled, _ := strconv.ParseBool(v)
	return enabled
}

func initializeGlobalWriter() {
	if p := os.Getenv(envSSLKey); p != "" {
		globalWriter = &keyLogWriter{
			pattern: p,
		}
		if verboseEnabled() {
			fmt.Fprintf(os.Stderr, "sslkeylogfile: enabled, writing TLS keys to %s\n", p)
		}
	}
}

func init() {
	initializeGlobalWriter()
}
