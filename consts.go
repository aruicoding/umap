package umap

import "unsafe"

// umap.go
const (
	bucketCnt     = 8
	loadFactorNum = 13
	loadFactorDen = 2

	bucketSize uintptr = unsafe.Sizeof(&ubmap{})

	emptyRest      = 0 // this cell is empty, and there are no more non-empty cells at higher indexes or overflows.
	emptyOne       = 1 // this cell is empty
	evacuatedX     = 2 // key/elem is valid.  Entry has been evacuated to first half of larger table.
	evacuatedY     = 3 // same as above, but evacuated to second half of larger table.
	evacuatedEmpty = 4 // cell is empty, bucket is evacuated.
	minTopHash     = 5 // minimum tophash for a normal filled cell.

	// flags
	iterator     = 0b00000001 // there may be an iterator using buckets
	oldIterator  = 0b00000010 // there may be an iterator using oldbuckets
	hashWriting  = 0b00000100 // a goroutine is writing to the map
	sameSizeGrow = 0b00001000 // the current map growth is to a new map of the same size
)

// bmap.go
const (
	dataOffset     = unsafe.Sizeof([bucketCnt]uint8{})
	overflowOffset = dataOffset + unsafe.Sizeof([bucketCnt * 2]any{})
	kvOffset       = unsafe.Sizeof(any(nil))
	keysOffset     = unsafe.Sizeof(any(nil)) * bucketCnt
)

// util
const (
	uintptrSize   = unsafe.Sizeof(uintptr(0)) * 8
	uintptrSize32 = 32
	uintptrSize64 = 64
	// uintptr 在64位机器上是8字节 在32位机器上是4字节 1个字节有8位
	// unsafe.Sizeof(uintptr(0))*8 -1 通过与操作后 1最多左移63位
	uintptrMax = unsafe.Sizeof(uintptr(0))*8 - 1
)
