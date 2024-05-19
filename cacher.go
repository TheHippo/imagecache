package imagecache

import "context"

// Cacher is the interface [Cache] expects to cache transformed images to.
type Cacher interface {
	Put(ctx context.Context, name string, content []byte) error
	Exists(ctx context.Context, name string) bool
	Delete(ctx context.Context, name string) error
	Get(ctx context.Context, name string) ([]byte, error)
	// Stats() (count int32, size int64)
}
