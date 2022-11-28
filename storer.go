package imagecache

type Storer interface {
	Exists(name string) bool
	Get(name string) ([]byte, error)
}
