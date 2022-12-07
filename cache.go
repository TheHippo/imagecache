package imagecache

import (
	"fmt"
	"net/http"

	"github.com/h2non/bimg"
)

// Cache caches items
type Cache struct {
	store  Storer
	layers []*Layer
}

// Creates a new Cache. Items are taken frrom the [Store]. Items are removed
// from the cache when certain criteria from each [Layer] are hit.
func New(store Storer, layers ...*Layer) *Cache {
	return &Cache{
		store:  store,
		layers: layers,
	}
}

func (c *Cache) putInCache(name string, content []byte, above int) {
	if above <= 0 {
		above = len(c.layers)
	}
	for i := 0; i < above; i++ {
		c.layers[i].Put(name, content) //nolint:errcheck
	}
}

// Clear a single item from the cache. There is no feedback on how successful the
// operation was and which layer produced an error.
func (c *Cache) Clear(name string) {
	for _, l := range c.layers {
		l.Delete(name) //nolint:errcheck
	}
}

type Handler func(string, http.ResponseWriter)

func (c *Cache) Handle(imageType bimg.ImageType, config bimg.Options) (Handler, error) {
	contentType, ctOk := contentTypes[imageType]
	if !SupportsType(imageType) || !ctOk {
		return nil, fmt.Errorf("image type %s is not supported", bimg.ImageTypeName(imageType))
	}

	cacheKey := fmt.Sprintf("%s-%s", bimg.ImageTypeName(imageType), cacheKey(config))

	return func(name string, w http.ResponseWriter) {
		w.Header().Set("Content-Type", contentType)

		// check if it is in one of the caches
		cacheName := fmt.Sprintf("%s-%s", cacheKey, name)
		for i, l := range c.layers {
			if !l.Exists(cacheName) {
				continue
			}

			content, err := l.Get(cacheName)
			if i > 0 && err == nil {
				// does not exist in higher up caches
				go c.putInCache(cacheName, content, i)
			}
			if err == nil {
				writeImage(w, content)
				return
			}
		}

		// not in cache

		// check if image exists
		if !c.store.Exists(name) {
			notFound(w)
			return
		}

		content, err := c.store.Get(name)
		if err != nil {
			// it should be there
			internalError(w)
			return
		}

		transformed, err := handleImage(content, config, imageType)
		if err != nil {
			internalError(w)
			return
		}
		// put in all caches
		go c.putInCache(cacheName, transformed, -1)

		writeImage(w, transformed)
	}, nil
}
