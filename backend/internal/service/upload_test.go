package service

import (
	"bytes"
	"strings"
	"testing"

	"gowiki/internal/config"
)

func newUploadSvc(t *testing.T, maxBytes int64) *UploadService {
	t.Helper()
	dir := t.TempDir()
	return NewUploadService(config.Config{UploadDir: dir, UploadMaxBytes: maxBytes})
}

func TestUploadSizeBoundary(t *testing.T) {
	const max = int64(8) * 1024 * 1024
	cases := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"zero rejected", 0, true},
		{"negative rejected", -1, true},
		{"just under max ok", max - 1, false},
		{"exactly max ok", max, false},
		{"one over max rejected", max + 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newUploadSvc(t, max)
			var data []byte
			if tc.size > 0 {
				data = bytes.Repeat([]byte("x"), int(tc.size))
			}
			_, err := svc.Save("a.png", tc.size, bytes.NewReader(data))
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected ok at size %d, got: %v", tc.size, err)
			}
		})
	}
}

func TestUploadRejectsBadExt(t *testing.T) {
	svc := newUploadSvc(t, int64(8)*1024*1024)
	_, err := svc.Save("a.exe", 10, strings.NewReader("xxxxxxxxxx"))
	if err == nil {
		t.Fatal("expected error for disallowed extension")
	}
}
