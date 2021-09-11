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
	return &SafePeer{
		BasicPeer: NewBasicPeer(gs).WithTag("safe"),
		lock:      &sync.Mutex{},
	}
}

// This hook can be safely triggered concurrently.
func (sp *SafePeer) OnWriteSafe(f func([]byte) error) *SafePeer {
	sp.onWrite = func(data []byte) error {
		sp.lock.Lock()
		defer sp.lock.Unlock()

		return f(data)
	}
	return sp
}

// This hook can be safely triggered concurrently.
func (sp *SafePeer) OnCloseSafe(f func() error) *SafePeer {
	sp.onClose = func() error {
		sp.lock.Lock()
		defer sp.lock.Unlock()

		return f()
	}
	return sp
}

// This hook can be safely triggered concurrently.
func (sp *SafePeer) OnStartSafe(f func()) *SafePeer {
	sp.onStart = func() {
		sp.lock.Lock()
		defer sp.lock.Unlock()
		f()
	}
	return sp
}

func (sp *SafePeer) Lock() {
	sp.lock.Lock()
}

func (sp *SafePeer) Unlock() {
	sp.lock.Unlock()
}
