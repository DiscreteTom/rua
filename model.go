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

type SmallestLogger interface {
	Print(v ...interface{})
}

// You can use `NewBasicSimpleLogger` to create a SimpleLogger.
type SimpleLogger interface {
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	Panic(v ...interface{})
}

// You can use `NewBasicLogger` to create a Logger.
type Logger interface {
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	Panic(v ...interface{})
	Tracef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
}
