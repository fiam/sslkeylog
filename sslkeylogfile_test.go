package sslkeylogfile

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const (
	testFilename = "sslkeylogfile.log"
)

func newTmpDir(t *testing.T) string {
	t.Helper()
	tmpdir, err := os.MkdirTemp("", "sslkeylogfile-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpdir)
	})
	return tmpdir
}

func TestKeyLogFile(t *testing.T) {
	t.Parallel()

	tmpdir := newTmpDir(t)

	keylogfile := filepath.Join(tmpdir, testFilename)
	klw := &keyLogWriter{pattern: keylogfile}
	w, err := klw.NewKeyWriter()
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				KeyLogWriter: w,
			},
		},
	}
	resp, err := client.Get("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	info, err := os.Stat(keylogfile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected keylogfile %q to be non-empty", keylogfile)
	}
}

func TestMultipleWriters(t *testing.T) {
	t.Parallel()
	const numWriters = 1000
	tmpdir := newTmpDir(t)
	keylogfile := filepath.Join(tmpdir, testFilename)
	klw := &keyLogWriter{pattern: keylogfile}
	errCh := make(chan error, numWriters)
	var wg sync.WaitGroup
	wg.Add(numWriters)
	for range numWriters {
		go func() {
			defer wg.Done()
			_, err := klw.NewKeyWriter()
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != numWriters {
		t.Fatalf("expected %d keylog files, got %d", numWriters, len(entries))
	}
}

func TestGlobalWriter(t *testing.T) { //nolint:paralleltest
	prev, isSet := os.LookupEnv(envSSLKey)
	if isSet {
		t.Cleanup(func() {
			_ = os.Setenv(envSSLKey, prev)
		})
	}
	if err := os.Unsetenv(envSSLKey); err != nil {
		t.Fatal(err)
	}
	globalWriter = nil
	initializeGlobalWriter()
	if globalWriter != nil {
		t.Fatal("expected globalWriter to be nil when env var is not set")
	}
	if err := os.Setenv(envSSLKey, "some-file"); err != nil {
		t.Fatal(err)
	}
	initializeGlobalWriter()
	if globalWriter == nil {
		t.Fatal("expected globalWriter to be non-nil when env var is set")
	}
}

func TestPatchDefaultTransport(t *testing.T) { //nolint:paralleltest
	tt, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatal("http.DefaultTransport is not an http.Transport")
	}
	tmpdir := newTmpDir(t)
	globalWriter = &keyLogWriter{pattern: filepath.Join(tmpdir, testFilename)}
	if err := PatchDefaultTransport(); err != nil {
		t.Fatal(err)
	}
	if tt.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig to be non-nil")
	}
	if tt.TLSClientConfig.KeyLogWriter == nil {
		t.Fatal("expected KeyLogWriter to be non-nil")
	}
}
