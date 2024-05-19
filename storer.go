package imagecache

import "context"

// Storer is the interface [Cache] expects to retrieve items from
type Storer interface {
	Exists(ctx context.Context, name string) bool
	Get(ctx context.Context, name string) ([]byte, error)
}
