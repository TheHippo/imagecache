package imagecache

// Storer is the interface [Cache] expects to retrieve items from
type Storer interface {
	Exists(name string) bool
	Get(name string) ([]byte, error)
}
