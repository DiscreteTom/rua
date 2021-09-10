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
	sp := &SafePeer{
		lock: &sync.Mutex{},
	}

	bp, err := NewBasicPeer(
		gs,
		Tag("safe"),
	)
	if err != nil {
		return nil, err
	}

	for _, o := range options {
		if err := o(bp); err != nil {
			return nil, err
		}
	}

	sp.BasicPeer = bp
	return sp, nil
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
