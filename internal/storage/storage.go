package storage

import (
	"context"
	"io/fs"
)

type Storage interface {
	Read(ctx context.Context, path string) ([]byte, error)
	WriteAtomic(ctx context.Context, path string, data []byte, mode fs.FileMode) error
	List(ctx context.Context, dir string) ([]fs.DirEntry, error)
	Delete(ctx context.Context, path string) error
	MkdirAll(ctx context.Context, dir string, mode fs.FileMode) error
}
