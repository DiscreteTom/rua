package kcp

import (
	"DiscreteTom/rua/pkg/model"
	"log"

	"github.com/xtaci/kcp-go/v5"
)

type kcpListener struct {
	addr    string
	gs      model.GameServer
	key     []byte
	bufSize int
}

func NewKcpListener(addr string, gs model.GameServer, key []byte, bufSize int) *kcpListener {
	return &kcpListener{
		addr:    addr,
		gs:      gs,
		key:     key,
		bufSize: bufSize,
	}
}

func (l *kcpListener) Start() error {
	log.Println("kcp server is listening at", l.addr)
	block, _ := kcp.NewAESBlockCrypt(l.key)
	listener, err := kcp.ListenWithOptions(l.addr, block, 10, 3)
	if err != nil {
		return err
	}
	for {
		c, err := listener.AcceptKCP()
		if err != nil {
			return err
		}
		p := &kcpPeer{c: c, gs: l.gs, bufSize: l.bufSize}
		l.gs.AddPeer(p)
	}
}
