package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer/debug"
	"github.com/DiscreteTom/rua/peer/persistent"
)

func main() {
	s, _ := rua.NewEventDrivenServer()
	s.OnPeerMsg(func(msg *rua.PeerMsg) {
		s.ForEachPeer(func(id int, peer rua.Peer) {
			if peer.Tag() == "file" {
				peer.Write(msg.Data)
			}
		})
	})

	sp, _ := debug.NewStdioPeer(s)
	s.AddPeer(sp)
	fp, _ := persistent.NewFilePeer("./log.txt", s)
	s.AddPeer(fp)

	s.Start()
}
