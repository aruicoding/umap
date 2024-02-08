package umap_test

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
	"umap"
	"unsafe"
)

func TestSizeof(t *testing.T) {
	t.Log(unsafe.Sizeof(any(nil)))
	t.Log(unsafe.Sizeof(any('k')))
	t.Log(unsafe.Sizeof(any("k")))
	t.Log(unsafe.Sizeof(any(0)))
	t.Log(unsafe.Sizeof(any(true)))
	t.Log(unsafe.Sizeof(0))
	t.Log(unsafe.Sizeof(true))
}

func TestAnd(t *testing.T) {
	var flag uint32
	const (
		iterator     uint32 = 0b00000001 // there may be an iterator using buckets
		oldIterator         = 0b00000010 // there may be an iterator using oldbuckets
		hashWriting         = 0b00000100 // a goroutine is writing to the map
		sameSizeGrow        = 0b00001000 // the current map growth is to a new map of the same size
	)
	t.Log(flag&iterator == iterator)
	t.Log(flag&oldIterator == oldIterator)
	t.Log(flag&hashWriting == hashWriting)
	t.Log(flag&sameSizeGrow == sameSizeGrow)
	atomic.StoreUint32(&flag, hashWriting)
	t.Log(flag&iterator == iterator)
	t.Log(flag&oldIterator == oldIterator)
	t.Log(flag&hashWriting == hashWriting)
	t.Log(flag&sameSizeGrow == sameSizeGrow)
	atomic.StoreUint32(&flag, 0)
	t.Log(flag&iterator == iterator)
	t.Log(flag&oldIterator == oldIterator)
	t.Log(flag&hashWriting == hashWriting)
	t.Log(flag&sameSizeGrow == sameSizeGrow)
}

func TestHash(t *testing.T) {
	hasher := fnv.New32a()
	seed := rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
	hasher.Write([]byte(fmt.Sprintf("%v", "aaa")))
	hasher.Write([]byte(strconv.FormatUint(uint64(seed), 10)))
	t.Log(hasher.Sum32())
}

func TestUmapHasher(t *testing.T) {
	// 2587109345
	// 2587109345
	m := umap.New(10)
	fmt.Println(m)

	for i := 0; i < 50; i++ {
		m.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		m.Get(fmt.Sprintf("key%d", i))
	}
	fmt.Println(m)
	// m.Get("id")
	// m.Get("name")
}

func BenchmarkUmapSetGet(b *testing.B) {
	m := umap.New(1000)

	for t := 0; t < b.N; t++ {
		for i := 0; i < 1000; i++ {
			m.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
			m.Get(fmt.Sprintf("key%d", i))
		}
	}
}

func BenchmarkMapSetGet(b *testing.B) {
	m := make(map[string]string, 1000)
	for i := 0; i < b.N; i++ {
		for i := 0; i < 1000; i++ {
			m[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
			_ = m[fmt.Sprintf("key%d", i)]
		}
	}
}

func TestMod(t *testing.T) {
	num1 := 7
	mod := 32
	t.Log(num1 & (mod - 1))
}

// bucketShift returns 1<<b, optimized for code generation.
func bucketShift(b uint8) uintptr {
	// Masking the shift amount allows overflow checks to be elided.
	return uintptr(1) << (b & (63 - 1))
}

// bucketMask returns 1<<b - 1, optimized for code generation.
func bucketMask(b uint8) uintptr {
	return bucketShift(b) - 1
}

func TestUintptrSize(t *testing.T) {
	t.Log(unsafe.Sizeof(uintptr(0)))
	t.Log(unsafe.Sizeof(uint8(0)))
	t.Logf("%032b", bucketMask(8))
}
