package kcp

import (
	"log"
	"time"

	"github.com/DiscreteTom/rua"

	"github.com/xtaci/kcp-go/v5"
)

type kcpPeer struct {
	id      int // peer id
	c       *kcp.UDPSession
	gs      rua.GameServer
	bufSize int
	timeout int
}

func (p *kcpPeer) Activate(id int) {
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
	p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond))
	for {
		buf := make([]byte, p.bufSize)
		n, err := p.c.Read(buf)
		if err != nil {
			if err.Error() == "timeout" {
				log.Printf("peer[%d] timeout\n", p.id)
			}
			p.gs.RemovePeer(p.id)
			break
		}
		p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond))

		p.gs.AppendPeerMsg(p.id, buf[:n])
	}
}
