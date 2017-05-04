package roulette

import (
	"bytes"
	"sync"
)

// wrapper over a buffer pool
type bufferPool struct {
	sp sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		sp: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (pool *bufferPool) get() *bytes.Buffer {
	return pool.sp.Get().(*bytes.Buffer)
}
func (pool *bufferPool) put(buffer *bytes.Buffer) {
	buffer.Reset()
	pool.sp.Put(buffer)
}
