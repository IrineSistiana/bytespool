package bytespool

import (
	"fmt"
	"math/bits"
	"sync"
	"unsafe"
)

// Note: The maximum reusable size is 2^30 (1GBytes).
// If size is too big, Get simply calls make([]byte,size) and
// ReleaseBuf is a noop.
func Get(size int) []byte {
	return _globalPool.get(size)
}

func Release(b []byte) {
	_globalPool.release(b)
}

var _globalPool = pool{}

type pool struct {
	sp smallPool
	lp largePool
}

func (p *pool) get(size int) []byte {
	switch {
	case size <= 0:
		return []byte{}
	case size <= 1<<8:
		return p.sp.get(size)
	case size <= 1<<30:
		return p.lp.get(size)
	default:
		return make([]byte, size)
	}
}

func (p *pool) release(b []byte) {
	if b == nil {
		panic("releasing a nil []byte")
	}

	c := cap(b)
	switch {
	case c == 0:
		return
	case c <= 1<<8:
		p.sp.release(b)
		return
	case c <= 1<<30:
		p.lp.release(b)
		return
	default:
		return
	}
}

// 1~2^8(256)
type smallPool struct {
	arrayPools [9]sync.Pool
}

func spIdx(size int) int {
	return bits.Len(uint(size - 1))
}

func spSize(idx int) int {
	return 1 << idx
}

func (p *smallPool) get(size int) []byte {
	i := spIdx(size)
	arrayCap := spSize(i)
	array, _ := p.arrayPools[i].Get().(*byte)
	if array == nil {
		return make([]byte, size, arrayCap)
	}
	b := unsafe.Slice(array, arrayCap)
	return b[:size]
}

func (p *smallPool) release(b []byte) {
	c := cap(b)
	i := spIdx(c)
	if c != spSize(i) {
		panic(fmt.Sprintf("invalid cap %d for pool #%d", c, i))
	}
	array := unsafe.SliceData(b)
	p.arrayPools[i].Put(array)
}

// 2^8(257) ~ 2^30
type largePool struct {
	arrayPools [22][4]sync.Pool
}

func lpIdx(size int) (int, int) {
	b := bits.Len(uint(size - 1))
	l := ((size - 1) >> (b - 3)) & 0b11
	h := b - 9
	return h, l
}

func lpSize(h, l int) int {
	return 1<<(h+8) + (l+1)<<(h+6)
}

func (p *largePool) get(size int) []byte {
	h, l := lpIdx(size)
	arrayCap := lpSize(h, l)
	array, _ := p.arrayPools[h][l].Get().(*byte)
	if array == nil {
		return make([]byte, size, arrayCap)
	}
	b := unsafe.Slice(array, arrayCap)
	return b[:size]
}

func (p *largePool) release(b []byte) {
	c := cap(b)
	h, l := lpIdx(c)
	if c != lpSize(h, l) {
		panic(fmt.Sprintf("invalid cap %d for pool #%d.%d", c, h, l))
	}
	array := unsafe.SliceData(b)
	p.arrayPools[h][l].Put(array)
}
