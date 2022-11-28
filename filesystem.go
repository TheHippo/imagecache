package imagecache

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

func (fs *FileSystem) Get(name string) ([]byte, error) {
	return fs.nfs.Get(name)
}

func (fs *FileSystem) Delete(name string) error {
	return fs.nfs.Delete(name)
}

func (fs *FileSystem) Exists(name string) bool {
	return fs.nfs.Exists(name)
}

func (fs *FileSystem) Put(name string, data []byte) error {
	return fs.nfs.Put(name, data)
}

// func (f *FileSystem) Stats() (count int32, size int64) {
// 	return f.nfs.Stats()
// }
