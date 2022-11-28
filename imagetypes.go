package imagecache

import (
	"sync"

	"github.com/h2non/bimg"
	"golang.org/x/exp/slices"
)

var allTypes = []bimg.ImageType{
	bimg.JPEG,
	bimg.WEBP,
	bimg.PNG,
	bimg.TIFF,
	bimg.GIF,
	bimg.PDF,
	bimg.SVG,
	bimg.MAGICK,
	bimg.HEIF,
	bimg.AVIF,
}

var contentTypes = map[bimg.ImageType]string{
	bimg.JPEG: "image/jpeg",
	bimg.WEBP: "image/webp",
	bimg.PNG:  "image/png",
	bimg.TIFF: "image/tiff",
	bimg.GIF:  "image/gif",
	// bimg.PDF:  "application/pdf",
	// bimg.SVG:  "image/svg+xml",
	// bimg.MAGICK:"",
	bimg.HEIF: "image/heif",
	bimg.AVIF: "image/avif",
}

var supportedTypes []bimg.ImageType
var getOnce sync.Once

func GetSupportedTypes() []bimg.ImageType {
	getOnce.Do(func() {
		supportedTypes = make([]bimg.ImageType, 0, len(allTypes))
		for _, t := range allTypes {
			if bimg.IsTypeSupported(t) {
				supportedTypes = append(supportedTypes, t)
			}
		}
	})
	return supportedTypes
}

func SupportsType(t bimg.ImageType) bool {
	return slices.Contains(GetSupportedTypes(), t)
}
