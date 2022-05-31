package gitfs

import (
	"errors"
	"io"
	"io/fs"
	"time"
)

type gitFSDir struct {
	name    string
	entries []string
	offset  int
}

func (d *gitFSDir) Stat() (fs.FileInfo, error) {
	return &gitDirEntry{name: d.name}, nil
}

func (d *gitFSDir) ReadDir(count int) ([]fs.DirEntry, error) {
	n := len(d.entries) - d.offset
	if n == 0 {
		if count <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	if count > 0 && n > count {
		n = count
	}

	list := make([]fs.DirEntry, 0, n)
	for i := 0; i < n; i++ {
		name := d.entries[d.offset]
		list = append(list, &gitDirEntry{name: name})
		d.offset++
	}

	return list, nil
}

func (d *gitFSDir) Read(_ []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("is a directory")}
}

func (d *gitFSDir) Close() error {
	return nil
}

type gitDirEntry struct {
	name string
}

func (e *gitDirEntry) Name() string               { return e.name }
func (e *gitDirEntry) Size() int64                { return 0 }
func (e *gitDirEntry) Mode() fs.FileMode          { return fs.ModeDir }
func (e *gitDirEntry) ModTime() time.Time         { return time.Time{} }
func (e *gitDirEntry) IsDir() bool                { return e.Mode().IsDir() }
func (e *gitDirEntry) Sys() any                   { return nil }
func (e *gitDirEntry) Type() fs.FileMode          { return fs.ModeDir }
func (e *gitDirEntry) Info() (fs.FileInfo, error) { return e, nil }
