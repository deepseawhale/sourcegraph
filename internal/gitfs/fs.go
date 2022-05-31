package gitfs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type gitFS struct {
	rev    string
	dir    string
	prefix string
}

func New(rev, dir, prefix string) fs.FS {
	return &gitFS{rev, dir, prefix}
}

var gitTreePattern = regexp.MustCompile("^tree .+:.+\n")

func (fs *gitFS) Open(name string) (fs.File, error) {
	if name == "." || name == "/" {
		name = ""
	}
	revPath := makeRevPath(fs.rev, filepath.Join(fs.prefix, name))

	cmd := exec.Command("git", "show", revPath)
	cmd.Dir = fs.dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if ok := gitTreePattern.Match(out); ok {
		lines := bytes.Split(out, []byte("\n"))

		entries := make([]string, 0, len(lines))
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}

			entries = append(entries, strings.TrimRight(string(line), string(os.PathSeparator)))
		}

		return &gitFSDir{
			name:    name,
			entries: entries,
		}, nil
	}

	return &gitFSFile{
		name:       name,
		ReadCloser: io.NopCloser(bytes.NewReader(out)),
	}, nil
}

func makeRevPath(rev, path string) string {
	return fmt.Sprintf("%s^:%s", rev, path)
}
