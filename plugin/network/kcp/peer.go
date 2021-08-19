package kcp

import (
	"DiscreteTom/rua/pkg/model"

	"github.com/xtaci/kcp-go/v5"
)

type kcpPeer struct {
	rc      chan *model.PeerMsg // receiver channel
	id      int                 // peer id
	c       *kcp.UDPSession
	gs      model.GameServer
	bufSize int
}

func (p *kcpPeer) Activate(rc chan *model.PeerMsg, id int) {
	p.rc = rc
	p.id = id
}

func (p *kcpPeer) Write(data []byte) error {
	_, err := p.c.Write(data)
	return err
}

func (p *kcpPeer) Close() error {
	return p.c.Close() // close kcp conn
}

func (p *kcpPeer) Start() {
	for {
		buf := make([]byte, p.bufSize)
		n, err := p.c.Read(buf)
		if err != nil {
			p.gs.RemovePeer(p.id)
			break
		}

		p.rc <- &model.PeerMsg{PeerId: p.id, Data: buf[:n]}
	}
}
