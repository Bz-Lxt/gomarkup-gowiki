package service_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gowiki/internal/config"
	"gowiki/internal/service"
)

type coordinatedUploadReader struct {
	fill   byte
	before <-chan struct{}
	filled chan<- struct{}
	after  <-chan struct{}
	done   bool
}

func (r *coordinatedUploadReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	if r.before != nil {
		<-r.before
	}
	n := len(p)
	if n > 32*1024 {
		n = 32 * 1024
	}
	for i := 0; i < n; i++ {
		p[i] = r.fill
	}
	r.done = true
	close(r.filled)
	if r.after != nil {
		<-r.after
	}
	return n, nil
}

func TestUploadServiceConcurrentSavesKeepContentsIsolated(t *testing.T) {
	dir := t.TempDir()
	svc := service.NewUploadService(config.Config{
		UploadDir:      dir,
		UploadMaxBytes: 64 * 1024,
	})

	aFilled := make(chan struct{})
	bFilled := make(chan struct{})
	readers := []io.Reader{
		&coordinatedUploadReader{fill: 'a', filled: aFilled, after: bFilled},
		&coordinatedUploadReader{fill: 'b', before: aFilled, filled: bFilled},
	}

	urls := make([]string, len(readers))
	errs := make([]error, len(readers))
	var wg sync.WaitGroup
	for i := range readers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			urls[i], errs[i] = svc.Save("avatar.png", 32*1024, readers[i])
		}(i)
	}
	wg.Wait()

	for i, want := range []byte{'a', 'b'} {
		if errs[i] != nil {
			t.Fatalf("save %d: %v", i, errs[i])
		}
		got, err := os.ReadFile(filepath.Join(dir, strings.TrimPrefix(urls[i], "/uploads/")))
		if err != nil {
			t.Fatalf("read saved upload %d: %v", i, err)
		}
		if !bytes.Equal(got, bytes.Repeat([]byte{want}, 32*1024)) {
			t.Fatalf("upload %d contains bytes from another concurrent request", i)
		}
	}
}
