package imagecache

import "context"

type FileSystem struct {
	nfs *NestedFileSystem
}

// compile-time check
var _ Cacher = &FileSystem{}
var _ Storer = &FileSystem{}

func NewFileSystem(path string) (*FileSystem, error) {
	nfs, err := NewNestedFilesystem(path, 0)
	return &FileSystem{
		nfs: nfs,
	}, err

}

func (fs *FileSystem) Get(ctx context.Context, name string) ([]byte, error) {
	return fs.nfs.Get(ctx, name)
}

func (fs *FileSystem) Delete(ctx context.Context, name string) error {
	return fs.nfs.Delete(ctx, name)
}

func (fs *FileSystem) Exists(ctx context.Context, name string) bool {
	return fs.nfs.Exists(ctx, name)
}

func (fs *FileSystem) Put(ctx context.Context, name string, data []byte) error {
	return fs.nfs.Put(ctx, name, data)
}

// func (f *FileSystem) Stats() (count int32, size int64) {
// 	return f.nfs.Stats()
// }
