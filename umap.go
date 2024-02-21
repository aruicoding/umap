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
	if cap <= bucketCnt {
		u.buckets = makeBucketArray(0)
		return u
	}
	cnt := 1 << u.b * bucketCnt
	for cap > uint(cnt) {
		u.b++
		cnt = 1 << u.b * bucketCnt
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
	// if !u.growing() && (overLoadFactor(u.count, u.b) || tooManyOverflowBuckets(u.noverflow, u.b)) {
	// 	log.Printf("Set(%v): hashGrow.\n", k)
	// 	u.hashGrow()
	// }
	if !u.growing() && overLoadFactor(u.count, u.b) {
		// log.Printf("Set(%v): hashGrow.\n", k)
		u.hashGrow()
	}
	hash := u.hasher(k)
	bindex := hash & bucketMask(u.b)
	bucket := *(**ubmap)(unsafe.Add(u.buckets, bucketSize*bindex))
	if u.growing() {
		// log.Printf("Set(%v): growWork.\n", k)
		u.growWork(bindex)
	}
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

	if u.growing() {
		// log.Println("Get(): growing.")
		bindex := bindex & u.oldbucketMask()
		bucket := *(**ubmap)(unsafe.Add(u.oldbuckets, bucketSize*bindex))
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
				// log.Println("Get(): found in oldbucket.")
			}
		}
	}

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
	bindex := hash & bucketMask(u.b)
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
	var it []bool
	u.flags = iterator
	if u.growing() {
		it = make([]bool, bucketShift(u.b-1))
	}
	for i := uintptr(0); i < bucketShift(u.b); i++ {
		bucket := *(**ubmap)(unsafe.Add(u.buckets, bucketSize*i))
		if u.growing() {
			u.flags = oldIterator
			oi := i & u.oldbucketMask()
			bucket := *(**ubmap)(unsafe.Add(u.oldbuckets, bucketSize*oi))
			if !it[oi] && !bucket.evacuated() {
				it[oi] = true
				for i := uintptr(0); i <= bucketCnt; i++ {
					if i == bucketCnt {
						bucket = bucket.overflow()
						if bucket == nil {
							break
						}
						i = 0
					}
					k := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i))
					v := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i))
					if k != any(nil) {
						fn(k, v)
					}
				}
			}
		}
		for i := uintptr(0); i <= bucketCnt; i++ {
			if i == bucketCnt {
				bucket = bucket.overflow()
				if bucket == nil {
					break
				}
				i = 0
			}
			k := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i))
			v := *(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i))
			if k != any(nil) {
				fn(k, v)
			}
		}
	}
}

func (u *umap) incrnoverflow() {
	if u.b < 16 {
		u.noverflow++
		return
	}
	if rand.Intn(10)%2 == 0 {
		u.noverflow++
	}
}

func (u *umap) hashGrow() {
	bigger := uint8(1)
	// if tooManyOverflowBuckets(u.noverflow, u.b) {
	// 	log.Println("hashGrow: sameSizeGrow")
	// 	bigger = uint8(0)
	// 	u.flags = sameSizeGrow
	// 	u.hash0 = hash0()
	// }
	u.b += bigger
	u.oldbuckets = u.buckets
	u.buckets = makeBucketArray(u.b)
	// log.Printf("%+v\n", u)
	u.noverflow = 0
}

func (u *umap) growing() bool {
	return u.oldbuckets != nil
}

func (u *umap) oldbucketMask() uintptr {
	b := u.b - 1
	if u.flags&sameSizeGrow != 0 {
		b = u.b
	}
	return uintptr(1)<<(uintptr(b)&uintptrMax) - 1
}

func (u *umap) growWork(bucket uintptr) {
	u.evacuate(bucket & u.oldbucketMask())
	if u.growing() {
		u.evacuate(u.nevacuate)
	}
}

func (u *umap) evacuate(oldbucket uintptr) {
	b := *(**ubmap)(unsafe.Add(u.oldbuckets, bucketSize*oldbucket))
	if !b.evacuated() {
		b.tophash[0] = evacuatedEmpty
		for ; b != nil; b = b.overflow() {
			for i := uintptr(0); i < bucketCnt; i++ {
				k := *(*any)(unsafe.Add(unsafe.Pointer(b), dataOffset+kvOffset*i))
				v := *(*any)(unsafe.Add(unsafe.Pointer(b), dataOffset+keysOffset+kvOffset*i))
				if k == any(nil) {
					continue
				}
				hash := u.hasher(k)
				bindex := hash & bucketMask(u.b)
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
							break
						}
					}
					if bucket.tophash[i] > 0 {
						continue
					}
					bucket.tophash[i] = top
					*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+kvOffset*i)) = k
					*(*any)(unsafe.Add(unsafe.Pointer(bucket), dataOffset+keysOffset+kvOffset*i)) = v
					break
				}
			}
		}
	}
	if oldbucket == u.nevacuate {
		u.nevacuate++
	}
	if u.nevacuate == bucketShift(u.b-1) {
		u.nevacuate = 0
		u.oldbuckets = nil
		if u.flags&sameSizeGrow != 0 {
			u.flags = 0
		}
	}
}
