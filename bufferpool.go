package roulette

import (
	"bytes"
	"sync"
)

// wrapper over a bytes buffer pool
type bytesPool struct {
	sp sync.Pool
}

func newBytesPool() *bytesPool {
	return &bytesPool{
		sp: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (pool *bytesPool) get() *bytes.Buffer {
	return pool.sp.Get().(*bytes.Buffer)
}
func (pool *bytesPool) put(buffer *bytes.Buffer) {
	buffer.Reset()
	pool.sp.Put(buffer)
}

// wrapper over a map[string]interface{} pool
type mapPool struct {
	sp sync.Pool
}

func newMapPool() *mapPool {
	return &mapPool{
		sp: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}
}

func (pool *mapPool) get() map[string]interface{} {
	return pool.sp.Get().(map[string]interface{})
}
func (pool *mapPool) put(buffer map[string]interface{}) {
	pool.sp.Put(buffer)
}

func (pool *mapPool) putReset(buffer map[string]interface{}) {
	for k := range buffer {
		delete(buffer, k)
	}
	pool.sp.Put(buffer)
}
