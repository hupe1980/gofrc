package gofrc

import (
	"sync"
	"time"
)

func RandomUint32() uint32 {
	r := rngPool.Get().(*rng)
	defer rngPool.Put(r)
	x := r.Uint32()

	return x
}

type rng struct {
	x uint32
}

var rngPool = sync.Pool{
	New: func() any {
		return &rng{x: getRandomUint32()}
	},
}

func (r *rng) Uint32() uint32 {
	r.x ^= r.x << 13
	r.x ^= r.x >> 17
	r.x ^= r.x << 5

	return r.x
}

func getRandomUint32() uint32 {
	x := time.Now().UnixNano()
	return uint32((x >> 32) ^ x)
}
