package umap

import (
	"fmt"
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
	fmt.Println(b)
	*(**ubmap)(unsafe.Add(unsafe.Pointer(b), overflowOffset)) = ovf
}

func (b *ubmap) keys() unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(b), dataOffset)
}
