package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer/debug"
	"github.com/DiscreteTom/rua/peer/network"
	"github.com/DiscreteTom/rua/peer/persistent"
)

func main() {
	s := rua.NewEventDrivenServer()
	mp := network.NewMultiplexPeer(s)

	s.AfterAddPeer(func(newPeer rua.Peer) {
		if newPeer.Tag() == "file" {
			mp.AddTarget(newPeer.Id(), newPeer)
		}
	}).AfterRemovePeer(func(targetId int) {
		mp.RemoveTarget(targetId)
	}).OnPeerMsg(func(m *rua.PeerMsg) {
		mp.Write(m.Data)
	})

	s.AddPeer(debug.NewStdioPeer(s))
	s.AddPeer(persistent.NewFilePeer("./output-1.txt", s))
	s.AddPeer(persistent.NewFilePeer("./output-2.txt", s))
	s.AddPeer(persistent.NewFilePeer("./output-3.txt", s))

	s.Start()
}
