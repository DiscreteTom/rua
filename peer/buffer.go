package peer

import (
	"time"

	"github.com/DiscreteTom/rua"
)

type BufferPeer struct {
	*SafePeer
	bufferSize   int
	queue        chan []byte
	writeTimeout int // write timeout in ms
	consumer     func(b []byte)
}

// Create a new BufferPeer.
// Data write to a BufferPeer will be stored in a buffer.
// You can use `WithConsumer` to register a consumer to consume those data.
func NewBufferPeer(gs rua.GameServer) *BufferPeer {
	bp := &BufferPeer{
		SafePeer:     NewSafePeer(gs),
		bufferSize:   256,
		writeTimeout: 1000,
		consumer:     func(b []byte) {},
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

func (bp *BufferPeer) WithConsumer(f func(b []byte)) *BufferPeer {
	bp.consumer = f
	return bp
}

func (bp *BufferPeer) OnStart(f func()) *BufferPeer {
	bp.SafePeer.OnStart(onStartFuncWrapper(bp, f))
	return bp
}

func (bp *BufferPeer) OnStartSafe(f func()) *BufferPeer {
	bp.SafePeer.OnStartSafe(onStartFuncWrapper(bp, f))
	return bp
}

func onStartFuncWrapper(bp *BufferPeer, f func()) func() {
	return func() {
		bp.queue = make(chan []byte, bp.bufferSize)
		go func() {
			for {
				data := <-bp.queue
				bp.consumer(data)
			}
		}()
		f()
	}
}
