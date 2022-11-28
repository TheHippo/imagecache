package imagecache

type Cacher interface {
	Put(name string, content []byte) error
	Exists(name string) bool
	Delete(name string) error
	Get(name string) ([]byte, error)
	// Stats() (count int32, size int64)
}
