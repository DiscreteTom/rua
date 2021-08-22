package rua

import "time"

type Peer interface {
	Activate(peerId int)
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
	RemovePeer(peerId int) error
	GetPeerCount() int
	AppendPeerMsg(peerId int, d []byte)
}

type GameServerLifeCycleEvent int

const (
	BeforeAddPeer GameServerLifeCycleEvent = iota
	AfterAddPeer
	BeforeRemovePeer
	AfterRemovePeer
	Step
	BeforeAppendPeerMsg
	AppendPeerMsg
)
