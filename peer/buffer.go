package peer

import (
	"time"

	"github.com/DiscreteTom/rua"
)

type BufferPeer struct {
	*SafePeer
	bufferSize    int
	queue         chan []byte
	writeTimeout  int // write timeout in ms
	consumer      func(b []byte) error
	onStartBuffer func()
}

// Create a new BufferPeer.
// Data write to a BufferPeer will be stored in a buffer.
// Use `WithConsumer` instead of `OnWrite` to register a consumer to consume those data.
// Use `OnStartBuffer` instead of `OnStart` to register the `onStart` hook for BufferPeer.
func NewBufferPeer(gs rua.GameServer) *BufferPeer {
	bp := &BufferPeer{
		SafePeer:      NewSafePeer(gs),
		bufferSize:    256,
		writeTimeout:  1000,
		consumer:      func(b []byte) error { return nil },
		onStartBuffer: func() {},
	}

	bp.SafePeer.
		OnWriteSafe(func(b []byte) error {
			timer := time.NewTimer(time.Duration(bp.writeTimeout) * time.Millisecond)
			select {
			case <-timer.C:
				return rua.ErrPeerWriteTimeout
			case bp.queue <- b:
				return nil
			}
		}).
		OnStart(func() {
			bp.queue = make(chan []byte, bp.bufferSize)
			go func() {
				for {
					data := <-bp.queue
					if err := bp.consumer(data); err != nil {
						bp.Logger().Errorf("rua.BufferPeer.Consume: %v", err)
					}
				}
			}()
			bp.onStartBuffer()
		}).
		WithTag("buffer")
	return bp
}

func (bp *BufferPeer) WithBufferSize(n int) *BufferPeer {
	bp.bufferSize = n
	return bp
}

func (bp *BufferPeer) WithWriteTimeout(ms int) *BufferPeer {
	bp.writeTimeout = ms
	return bp
}

// This consumer will NOT be triggered in parallel.
func (bp *BufferPeer) WithConsumer(f func(b []byte) error) *BufferPeer {
	bp.consumer = f
	return bp
}

func (bp *BufferPeer) OnStartBuffer(f func()) *BufferPeer {
	bp.onStartBuffer = f
	return bp
}
