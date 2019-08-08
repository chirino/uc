package files

import (
	"fmt"
	"github.com/chirino/uc/internal/pkg/utils"
	"io"
	"os"
	"path/filepath"
)

func WithCreateThenReplace(fileName string, perm os.FileMode, action func(*os.File) error) error {
	dir := filepath.Dir(fileName)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	base := filepath.Base(fileName)
	newName := filepath.Join(dir, fmt.Sprintf(".%s.new", base))
	oldName := filepath.Join(dir, fmt.Sprintf(".%s.old", base))

	f, err := os.OpenFile(newName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close() // in case of a panic in action..
	if err := action(f); err != nil {
		return err
	}
	f.Close() // make sure we close it eagerly or else Replace wont work right on windows..
	return Replace(fileName, newName, oldName)
}

func Replace(target string, src string, backupPath string) error {

	// clean up any previous backup attempt..
	_ = os.Remove(backupPath)

	// If the target does not exist, we can take a shortcut...
	if _, err := os.Stat(target); err != nil {
		return os.Rename(src, target)
	}

	// target exists, move it out the way and into the backup...
	if err := os.Rename(target, backupPath); err != nil {
		return err
	}

	// Lets see if we can replace the target.
	if err1 := os.Rename(src, target); err1 != nil {
		// restore the backup...
		if err2 := os.Rename(backupPath, target); err2 != nil {
			return utils.Errors(err1, err2)
		}
		return err1
	}

	// we don't need the backup any more.. this can fail on windows if the
	// original target file (moved to the backup) was still open.
	_ = os.Remove(backupPath)

	return nil
}

func WithOpenFile(name string, flag int, perm os.FileMode, action func(*os.File) error) error {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	return action(f)
}

func WithFileReader(from string, action func(io.Reader) error, filters ...func(r io.Reader) (io.Reader, error)) error {
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer source.Close()
	return WithReader(source, action, filters...)
}

func WithReader(source io.Reader, action func(io.Reader) error, filters ...func(r io.Reader) (io.Reader, error)) (err error) {
	reader := io.Reader(source)
	for _, f := range filters {
		reader, err = f(reader)
		if err != nil {
			return err
		}
		if reader, ok := reader.(io.Closer); ok {
			defer reader.Close()
		}
	}
	return action(reader)
}
