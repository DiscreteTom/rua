package network

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type MultiplexPeer struct {
	*peer.SafePeer
	targets  map[int]rua.Peer
	sync     bool
	parallel bool
}

// Create a new multiplex peer.
// Use `WithSyncWrite` and `WithParallelWrite` to change the write behavior.
// Use `AddTarget` and `RemoveTarget` to set targets.
func NewMultiplexPeer(gs rua.GameServer) *MultiplexPeer {
	mp := &MultiplexPeer{
		SafePeer: peer.NewSafePeer(gs),
		targets:  map[int]rua.Peer{},
		sync:     false,
		parallel: true,
	}
	mp.SafePeer.
		OnWrite(func(b []byte) error {
			work := func() {
				writeOrLog := func(p rua.Peer, b []byte) {
					if err := p.Write(b); err != nil {
						mp.Logger().Error("rua.MultiplexPeer.Write:", err)
					}
				}
				mp.Lock()
				defer mp.Unlock()
				for _, p := range mp.targets {
					if mp.parallel {
						go writeOrLog(p, b)
					} else {
						writeOrLog(p, b)
					}
				}
			}

			if mp.sync {
				work()
			} else {
				go work()
			}
			return nil
		}).
		WithTag("multiplex")

	return mp
}

func (mp *MultiplexPeer) WithSyncWrite(sync bool) *MultiplexPeer {
	mp.sync = sync
	return mp
}
func (mp *MultiplexPeer) WithParallelWrite(parallel bool) *MultiplexPeer {
	mp.parallel = parallel
	return mp
}

func (mp *MultiplexPeer) AddTarget(targetId int, p rua.Peer) {
	mp.Lock()
	defer mp.Unlock()

	mp.targets[targetId] = p
}

func (mp *MultiplexPeer) RemoveTarget(targetId int) {
	mp.Lock()
	defer mp.Unlock()

	delete(mp.targets, targetId)
}
