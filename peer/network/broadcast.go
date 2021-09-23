package network

import (
	"sync"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type BroadcastPeer struct {
	*peer.SafePeer
	sync     bool
	selector func(p rua.Peer) bool
}

// Create a new BroadcastPeer.
// By default, the broadcast peer will broadcast message in parallel to all other peers except it self.
// You can use `WithSelector` to select broadcast targets.
// The BroadcastPeer will never write message to itself to avoid recursive call.
func NewBroadcastPeer(gs rua.GameServer) *BroadcastPeer {
	bp := &BroadcastPeer{
		SafePeer: peer.NewSafePeer(gs),
		sync:     false,
		selector: func(p rua.Peer) bool { return true },
	}

	bp.SafePeer.
		OnWrite(func(b []byte) error {
			wg := sync.WaitGroup{}

			bp.GameServer().ForEachPeer(func(peer rua.Peer) {
				if bp.Id() != peer.Id() && bp.selector(peer) {
					wg.Add(1)
					go func(p rua.Peer) {
						rua.WriteOrLog(p, b)
						wg.Done()
					}(peer)
				}
			})
			if bp.sync {
				wg.Done()
			}
			return nil
		}).
		WithTag("broadcast")

	return bp
}

// If the selector returns true, the target peer will be notified.
func (bp *BroadcastPeer) WithSelector(f func(p rua.Peer) bool) *BroadcastPeer {
	bp.selector = f
	return bp
}

func (bp *BroadcastPeer) WithSyncWrite(sync bool) *BroadcastPeer {
	bp.sync = sync
	return bp
}
