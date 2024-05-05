package imagecache

import (
	"hash/fnv"
	"math/rand"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var litter_size = int64(len(letters))

func randSeq(r rand.Source, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Int63()%litter_size]
	}
	return string(b)
}

func BenchmarkFNV64(b *testing.B) {
	seed := time.Now().UnixNano()
	b.Run("new", func(b *testing.B) {
		r := rand.New(rand.NewSource(seed))
		h := fnv.New64a()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h.Write([]byte(randSeq(r, 200)))
			_ = h.Sum(nil)
			h.Reset()
		}
	})
	b.Run("reuse", func(b *testing.B) {
		r := rand.New(rand.NewSource(seed))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h := fnv.New64a()
			h.Write([]byte(randSeq(r, 200)))
			_ = h.Sum(nil)
		}
	})
}

func BenchmarkFNV32(b *testing.B) {
	seed := time.Now().UnixNano()
	b.Run("new", func(b *testing.B) {
		r := rand.New(rand.NewSource(seed))
		h := fnv.New32a()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h.Write([]byte(randSeq(r, 200)))
			_ = h.Sum(nil)
			h.Reset()
		}
	})
	b.Run("reuse", func(b *testing.B) {
		r := rand.New(rand.NewSource(seed))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h := fnv.New32a()
			h.Write([]byte(randSeq(r, 200)))
			_ = h.Sum(nil)
		}
	})
}
