package cache

import (
	"github.com/spaolacci/murmur3"
	"math"
	"sync"
)

type BloomFilter struct {
	mu   sync.RWMutex
	bits []uint64
	m    uint64
	k    uint64
}

func NewBloomFilter(itemSize int, falsePositiveRate float64) *BloomFilter {
	// m = -n * ln(p) / (ln2)^2
	// k = m / n * ln2
	if itemSize <= 0 {
		itemSize = 100_000
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01
	}

	m := uint64(math.Ceil(-float64(itemSize) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2)))
	k := uint64(math.Ceil(float64(m) / float64(itemSize) * math.Ln2))
	sz := (m + 63) / 64

	return &BloomFilter{
		bits: make([]uint64, sz),
		m:    m,
		k:    k,
	}
}

func (bf *BloomFilter) Add(item string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	h1, h2 := murmur3.Sum128([]byte(item))

	for i := uint64(0); i < bf.k; i++ {
		pos := (h1 + i*h2) % bf.m
		bf.bits[pos/64] |= 1 << (pos % 64)
	}
}

func (bf *BloomFilter) MayContain(item string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	h1, h2 := murmur3.Sum128([]byte(item))

	for i := uint64(0); i < bf.k; i++ {
		pos := (h1 + i*h2) % bf.m
		if (bf.bits[pos/64] & (1 << (pos % 64))) == 0 {
			return false
		}
	}
	return true
}
