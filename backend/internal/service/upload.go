package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"gowiki/internal/config"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/timeutil"
)

var allowedExt = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
}

type UploadService struct {
	cfg config.Config
}

func NewUploadService(cfg config.Config) *UploadService {
	return &UploadService{
		cfg: cfg,
	}
}

func (s *UploadService) Save(filename string, size int64, r io.Reader) (string, error) {
	if size <= 0 || size > s.cfg.UploadMaxBytes {
		return "", apperr.New(apperr.CodeValidation, 400, "文件大小超出限制")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedExt[ext]; !ok {
		return "", apperr.New(apperr.CodeValidation, 400, "仅支持 png/jpg/gif/webp")
	}
	if err := os.MkdirAll(s.cfg.UploadDir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s_%s%s", timeutil.Now().Format("20060102150405"), uuid.New().String()[:8], ext)
	dst := filepath.Join(s.cfg.UploadDir, name)
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	// Use io.Copy instead of io.CopyBuffer with a shared buffer. The old
	// code passed a single shared []byte (s.copyBuf) to every concurrent
	// request, so two uploads racing on the same buffer caused one
	// request's bytes to overwrite the other's in-flight data, producing
	// the observed "cross-talk" where the downloaded file content did not
	// match the originally uploaded PNG. io.Copy allocates a private
	// per-call buffer, eliminating the shared-state data race.
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}
