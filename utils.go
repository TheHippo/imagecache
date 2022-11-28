package imagecache

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/h2non/bimg"
)

func handleImage(content []byte, option bimg.Options, imageType bimg.ImageType) ([]byte, error) {
	image := bimg.NewImage(content)

	if _, err := image.Process(option); err != nil {
		return nil, err
	}

	if _, err := image.Convert(imageType); err != nil {
		return nil, err
	}
	return image.Image(), nil
}

func cacheKey(config bimg.Options) string {
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%+v", config)))
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

func internalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, "Internal error", http.StatusInternalServerError)
}

func notFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeImage(w http.ResponseWriter, content []byte) {
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.Write(content) //nolint:errcheck
}
