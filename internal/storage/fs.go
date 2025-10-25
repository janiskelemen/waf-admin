package storage

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yourname/waf-admin/internal/util"
)

type FS struct{}

func NewFS() *FS { return &FS{} }

func (FS) Read(_ context.Context, path string) ([]byte, error) { return os.ReadFile(path) }

func (FS) WriteAtomic(_ context.Context, path string, data []byte, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
	return util.AtomicWrite(path, data, mode)
}

func (FS) List(_ context.Context, dir string) ([]fs.DirEntry, error) { return os.ReadDir(dir) }

func (FS) Delete(_ context.Context, path string) error { return os.Remove(path) }

func (FS) MkdirAll(_ context.Context, dir string, mode fs.FileMode) error { return os.MkdirAll(dir, mode) }
