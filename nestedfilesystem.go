package imagecache

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"os"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"
const filePermission = 0776

type NestedFileSystem struct {
	path    string
	numSubs uint
}

// compile-time check
var _ Storer = &NestedFileSystem{}
var _ Cacher = &NestedFileSystem{}

func NewNestedFilesystem(path string, numSubdirectories uint) (*NestedFileSystem, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("'%s' is not a folder", path)
	}
	if numSubdirectories > uint(len(alphabet)) {
		return nil, fmt.Errorf("number of requested subdirectories is to large: %d (maximum: %d)", numSubdirectories, len(alphabet))
	}
	if numSubdirectories > 0 {
		var i uint
		for i = 0; i < numSubdirectories; i++ {
			subPath := fmt.Sprintf("%s/%s", path, string(alphabet[i%uint(len(alphabet))]))
			if err := os.Mkdir(subPath, filePermission); err != nil && !errors.Is(err, fs.ErrExist) {
				return nil, err
			}
		}
	}
	return &NestedFileSystem{
		path:    path,
		numSubs: numSubdirectories,
	}, nil
}

func (nfs *NestedFileSystem) calculatePath(name string) string {
	hash := fnv.New64a()

	hash.Write([]byte(name))
	hashedName := hash.Sum(nil)

	if nfs.numSubs > 0 {
		sum := hash.Sum64()
		return fmt.Sprintf("%s/%s/%x", nfs.path, string(alphabet[(uint(sum)%nfs.numSubs)%uint(len(alphabet))]), hashedName)
	}

	return fmt.Sprintf("%s/%x", nfs.path, hashedName)
}

func (nfs *NestedFileSystem) Put(_ context.Context, name string, content []byte) error {
	fn := nfs.calculatePath(name)
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermission)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(content)
	if err != nil {
		return err
	}
	if n != len(content) {
		return fmt.Errorf("expected %d bytes to be written, but only %d were", len(content), n)
	}

	return nil
}

func (nfs *NestedFileSystem) Get(_ context.Context, name string) ([]byte, error) {
	fn := nfs.calculatePath(name)
	f, err := os.OpenFile(fn, os.O_RDONLY, filePermission)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (nfs *NestedFileSystem) exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (nfs *NestedFileSystem) Exists(_ context.Context, name string) bool {
	return nfs.exists(nfs.calculatePath(name))
}

func (nfs *NestedFileSystem) Delete(_ context.Context, name string) error {
	return os.Remove(nfs.calculatePath(name))
}

// func (nfs *NestedFileSystem) Stats() (count int32, size int64) {
// 	filepath.Walk(nfs.path, func(_ string, info fs.FileInfo, err error) error { //nolint:errcheck
// 		if err == nil && !info.IsDir() {
// 			count++
// 			size += info.Size()
// 		}
// 		return nil
// 	})
// 	return
// }
