package service_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gowiki/internal/config"
	"gowiki/internal/service"
)

func TestUploadSaveAcceptsConfiguredSizeLimit(t *testing.T) {
	const limit = 16
	payload := bytes.Repeat([]byte{0x89}, limit)
	dir := t.TempDir()
	svc := service.NewUploadService(config.Config{
		UploadDir:      dir,
		UploadMaxBytes: limit,
	})

	url, err := svc.Save("boundary.png", int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Save() rejected an upload exactly at the configured limit: %v", err)
	}
	if !strings.HasPrefix(url, "/uploads/") {
		t.Fatalf("Save() URL = %q, want an /uploads/ URL", url)
	}
	got, err := os.ReadFile(filepath.Join(dir, filepath.Base(url)))
	if err != nil {
		t.Fatalf("read saved upload: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("saved content = %v, want %v", got, payload)
	}
}
