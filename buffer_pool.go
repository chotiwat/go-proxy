package proxy

import "sync"

// A BufferPool is an interface for getting and returning temporary
// byte slices for use by io.CopyBuffer.
type BufferPool interface {
	Get() []byte
	Put([]byte)
}

// NewBufferPool returns a new buffer pool.
func NewBufferPool() BufferPool {
	return &localBufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return []byte{}
			},
		},
	}
}

type localBufferPool struct {
	pool *sync.Pool
}

func (lbp *localBufferPool) Get() []byte {
	return lbp.pool.Get().([]byte)
}

func (lbp *localBufferPool) Put(buf []byte) {
	lbp.pool.Put(buf)
}
