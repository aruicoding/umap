package umap

import (
	"unsafe"
)

type ubmap struct {
	tophash [bucketCnt]uint8
	_       [bucketCnt * 2]any
	_       *ubmap
}

func (b *ubmap) overflow() *ubmap {
	return *(**ubmap)(unsafe.Add(unsafe.Pointer(b), overflowOffset))
}

func (b *ubmap) setoverflow(ovf *ubmap) {
	*(**ubmap)(unsafe.Add(unsafe.Pointer(b), overflowOffset)) = ovf
}

func (b *ubmap) evacuated() bool {
	return b.tophash[0] > emptyOne && b.tophash[0] < minTopHash
}
