package peer

import (
	"sync"

	"github.com/DiscreteTom/rua"
)

type SafePeer struct {
	*BasicPeer
	lock *sync.Mutex
}

func NewSafePeer(gs rua.GameServer) *SafePeer {
	sp := &SafePeer{
		lock: &sync.Mutex{},
	}

	bp := NewBasicPeer(gs)
	bp.SetTag("safe")

	sp.BasicPeer = bp
	return sp
}

// This hook can be safely triggered concurrently.
func (sp *SafePeer) OnWriteSafe(f func([]byte) error) {
	sp.onWrite = func(data []byte) error {
		sp.lock.Lock()
		defer sp.lock.Unlock()

		return f(data)
	}
}

// This hook can be safely triggered concurrently.
func (sp *SafePeer) OnCloseSafe(f func() error) {
	sp.onClose = func() error {
		sp.lock.Lock()
		defer sp.lock.Unlock()

		return f()
	}
}

// This hook can be safely triggered concurrently.
func (sp *SafePeer) OnStartSafe(f func()) {
	sp.onStart = func() {
		sp.lock.Lock()
		defer sp.lock.Unlock()
		f()
	}
}

func (sp *SafePeer) Lock() {
	sp.lock.Lock()
}

func (sp *SafePeer) Unlock() {
	sp.lock.Unlock()
}
