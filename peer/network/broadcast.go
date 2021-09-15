package network

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type BroadcastPeer struct {
	*peer.SafePeer
	sync   bool
	filter func(p rua.Peer) bool
}

// Create a new BroadcastPeer.
// By default, the broadcast peer will broadcast message to all other peers except it self.
// You can use `WithFilter` to change this hebavior.
func NewBroadcastPeer(gs rua.GameServer) *BroadcastPeer {
	bp := &BroadcastPeer{
		SafePeer: peer.NewSafePeer(gs),
		sync:     false,
	}
	bp.filter = func(p rua.Peer) bool { return p.Id() != bp.Id() }

	bp.SafePeer.
		OnWrite(func(b []byte) error {
			work := func(peer rua.Peer) {
				if bp.filter(peer) {
					if err := peer.Write(b); err != nil {
						bp.Logger().Error("rua.BroadcastPeer.Write:", err)
					}
				}
			}
			bp.GameServer().ForEachPeer(func(id int, peer rua.Peer) {
				if bp.sync {
					work(peer)
				} else {
					go work(peer)
				}
			})
			return nil
		}).
		WithTag("broadcast")

	return bp
}

// If filter return true, then the target peer will be broadcasted.
func (bp *BroadcastPeer) WithFilter(f func(p rua.Peer) bool) *BroadcastPeer {
	bp.filter = f
	return bp
}

func (bp *BroadcastPeer) WithSyncWrite(sync bool) *BroadcastPeer {
	bp.sync = sync
	return bp
}
