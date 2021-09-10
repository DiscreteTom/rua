package peer

import (
	"sync"

	"github.com/DiscreteTom/rua"
)

type SafePeer struct {
	*BasicPeer
	lock *sync.Mutex
}

func NewSafePeer(gs rua.GameServer, options ...BasicPeerOption) (*SafePeer, error) {
	p := &SafePeer{
		lock: &sync.Mutex{},
	}

	bp, err := NewBasicPeer(
		gs,
		Tag("safe"),
	)

	if err != nil {
		return nil, err
	}

	p.BasicPeer = bp
	return p, nil
}

// This hook can be safely triggered concurrently.
func (p *SafePeer) OnWriteSafe(f func([]byte) error) {
	p.onWrite = func(data []byte) error {
		p.lock.Lock()
		defer p.lock.Unlock()

		return f(data)
	}
}

// This hook can be safely triggered concurrently.
func (p *SafePeer) OnCloseSafe(f func() error) {
	p.onClose = func() error {
		p.lock.Lock()
		defer p.lock.Unlock()

		return f()
	}
}

// This hook can be safely triggered concurrently.
func (p *SafePeer) OnStartSafe(f func()) {
	p.onStart = func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		f()
	}
}
