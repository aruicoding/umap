package umap

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"unsafe"
)

func New(cap uint) *umap {
	u := &umap{
		hash0: hash0(),
	}
	if cap == 0 {
		return u
	}
	count := 1 << u.b * bucketCnt
	for cap > uint(count) {
		u.b++
		count = 1 << u.b * bucketCnt
	}
	u.buckets = makeBucketArray(u.b)
	return u
}

type umap struct {
	count      int
	flags      uintptr
	b          uint8
	noverflow  uint16
	hash0      uint32
	buckets    unsafe.Pointer
	oldbuckets unsafe.Pointer
	nevacuate  uintptr
}

func (u *umap) hasher(k any) uintptr {
	hasher := fnv.New32a()
	hasher.Write([]byte(fmt.Sprintf("%v", k)))
	hasher.Write([]byte(strconv.FormatUint(uint64(u.hash0), 10)))
	if uintptrSize == uintptrSize32 {
		return uintptr(hasher.Sum32())
	}
	if uintptrSize == uintptrSize64 {
		copyhash := uintptr(hasher.Sum32()) << 32
		return copyhash | uintptr(hasher.Sum32())
	}
	return uintptr(0)
}

func (u *umap) Len() int {
	return u.count
}

func (u *umap) Set(k any, v any) {
	u.count++
	if u.buckets == nil || overLoadFactor(u.count, u.b) || tooManyOverflowBuckets(u.noverflow, u.b) {
		u.hashGrow()
	}
	hash := u.hasher(k)
	bindex := hash & bucketMask((u.b))
	bucket := *(**ubmap)(unsafe.Add(u.buckets, bucketSize*bindex))
	top := tophash(hash)
	for i := uintptr(0); i <= bucketCnt; i++ {
		if i == bucketCnt {
			if bucket.overflow() == nil {
				ovf := new(ubmap)
				bucket.setoverflow(ovf)
				u.incrnoverflow()
			}
			bucket = bucket.overflow()
			i = 0
		}
		if bucket.tophash[i] == top {
			kk := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i))
			if keyEqual(&k, &kk) {
				*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i)) = v
				return
			}
		}
		if bucket.tophash[i] > 0 {
			continue
		}
		bucket.tophash[i] = top
		*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i)) = k
		*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i)) = v
		return
	}
}

func (u *umap) Get(k any) any {
	hash := u.hasher(k)
	bindex := hash & bucketMask((u.b))
	bucket := *(**ubmap)(unsafe.Add(u.buckets, bucketSize*bindex))
	top := tophash(hash)
	for i := uintptr(0); i <= bucketCnt; i++ {
		if i == bucketCnt {
			bucket = bucket.overflow()
			if bucket == nil {
				break
			}
			i = 0
		}
		if bucket.tophash[i] == top {
			kk := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i))
			if keyEqual(&k, &kk) {
				return *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i))
			}
		}
	}
	return any(nil)
}

func (u *umap) Del(k any) {
	hash := u.hasher(k)
	bindex := hash & bucketMask((u.b))
	bucket := *(**ubmap)(unsafe.Add(u.buckets, bucketSize*bindex))
	top := tophash(hash)
	for i := uintptr(0); i <= bucketCnt; i++ {
		if i == bucketCnt {
			bucket = bucket.overflow()
			if bucket == nil {
				break
			}
			i = 0
		}
		if bucket.tophash[i] == top {
			kk := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i))
			if keyEqual(&k, &kk) {
				*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i)) = any(nil)
				*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i)) = any(nil)
			}
			bucket.tophash[i] = 0
			return
		}
	}
}

func (u *umap) Iteration(fn callback) {
	// 获取k v
	// 调用callback
	fn(nil, nil)
}

func (u *umap) empty() bool {
	return u.buckets == nil
}

func (u *umap) incrnoverflow() {
	if u.b < 16 {
		u.noverflow++
		return
	}
	if rand.Intn(10)%2 == 1 {
		u.noverflow++
	}
}

func (u *umap) hashGrow() {
	bigger := uint8(1)
	u.b += bigger
	u.oldbuckets = u.buckets
	u.buckets = makeBucketArray(u.b)
}
