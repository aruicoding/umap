package umap

import (
	"math/rand"
	"reflect"
	"time"
	"unsafe"
)

func hash0() uint32 {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
}

func makeBucketArray(b uint8) unsafe.Pointer {
	s := make([]*ubmap, 1<<b)
	for i := 0; i < 1<<b; i++ {
		s[i] = new(ubmap)
	}
	// 数组首位指针就是数组指针
	return unsafe.Pointer(&s[0])
}

// 返回2^b 也就是桶个数
func bucketShift(b uint8) uintptr {
	return uintptr(1) << (uintptr(b) & uintptrMax)
}

func bucketMask(b uint8) uintptr {
	// uintptr(1) << b & uintptrMark 表示桶个数
	// 值 -1 就是掩码
	// eg. 当b为3时 00001000 掩码为00000111
	return uintptr(1)<<(uintptr(b)&uintptrMax) - 1
}

func tophash(hash uintptr) uint8 {
	top := uint8(hash >> (uintptrSize - 8))
	if top < minTopHash {
		top += minTopHash
	}
	return top
}

func keyEqual(k1 *any, k2 *any) bool {
	// return k1 == k2
	return reflect.DeepEqual(k1, k2)
}

func overLoadFactor(count int, b uint8) bool {
	return count > bucketCnt && uintptr(count) > loadFactorNum*(bucketShift(b)/loadFactorDen)
}

func tooManyOverflowBuckets(noverflow uint16, b uint8) bool {
	if b > 15 {
		b = 15
	}
	return noverflow >= uint16(bucketShift(b))
}
