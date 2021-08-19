package model

type Peer interface {
	Activate(chan *PeerMsg, int)
	Write([]byte) error
	Close() error
	Start()
}

type PeerMsg struct {
	PeerId int
	Data   []byte
}

type GameServer interface {
	AddPeer(Peer)
	RemovePeer(int) error
}
