package gitfs

import (
	"io"
	"io/fs"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type gitFSFile struct {
	name string
	size int64
	io.ReadCloser
}

func (f *gitFSFile) Stat() (fs.FileInfo, error) {
	return &gitFileEntry{name: f.name, size: f.size}, nil
}

func (d *gitFSFile) ReadDir(count int) ([]fs.DirEntry, error) {
	return nil, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("not a directory")}
}

type gitFileEntry struct {
	name string
	size int64
}

func (e *gitFileEntry) Name() string       { return e.name }
func (e *gitFileEntry) Size() int64        { return e.size }
func (e *gitFileEntry) Mode() fs.FileMode  { return fs.ModePerm }
func (e *gitFileEntry) ModTime() time.Time { return time.Time{} }
func (e *gitFileEntry) IsDir() bool        { return e.Mode().IsDir() }
func (e *gitFileEntry) Sys() any           { return nil }
