package rua

import "time"

type Peer interface {
	Activate(chan *PeerMsg, int)
	Write([]byte) error
	Close() error
	Start()
}

type PeerMsg struct {
	PeerId int
	Data   []byte
	Time   time.Time
}

type GameServer interface {
	AddPeer(Peer)
	RemovePeer(int) error
	GetPeerCount() int
}
