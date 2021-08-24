package rua

import "time"

type Peer interface {
	Activate(peerId int)
	Write([]byte) error
	Close() error
	Start() // start and wait
	GetId() int
	GetTag() string
	SetTag(string)
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
	Start() []error
	Stop()
}

type Logger interface {
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	Panic(v ...interface{})
}
