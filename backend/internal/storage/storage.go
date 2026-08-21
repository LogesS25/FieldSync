// Package storage implements local-disk file storage for uploads (currently
// just the daily handwritten report PDFs, business requirements §10). A
// real deployment should swap this for S3 or similar (see AGENTS.md
// production-readiness notes) — the interface is deliberately narrow so
// that swap doesn't ripple through the daily reports package.
package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func randomFilename() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate filename: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

type Storage struct {
	baseDir string
}

func New(baseDir string) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &Storage{baseDir: baseDir}, nil
}

// Save writes r to a new file under baseDir and returns the storage-relative
// path (safe to persist in the database), keyed by a generated UUID so
// student-supplied filenames never influence the on-disk path.
func (s *Storage) Save(subdir, ext string, r io.Reader) (relativePath string, err error) {
	dir := filepath.Join(s.baseDir, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage subdir: %w", err)
	}

	name, err := randomFilename()
	if err != nil {
		return "", err
	}
	relativePath = filepath.Join(subdir, name+ext)

	f, err := os.Create(filepath.Join(s.baseDir, relativePath))
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return relativePath, nil
}

// AbsolutePath resolves a stored relative path back to an absolute path on
// disk for serving.
func (s *Storage) AbsolutePath(relativePath string) string {
	return filepath.Join(s.baseDir, relativePath)
}
