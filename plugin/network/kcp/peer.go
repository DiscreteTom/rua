package kcp

import (
	"log"
	"sync"
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
	lock    sync.Mutex
	closed  bool
}

func (p *kcpPeer) Activate(id int) {
	p.id = id
	p.lock = sync.Mutex{}
	p.closed = false
}

// Thread safe.
func (p *kcpPeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := p.c.Write(data)
	return err
}

func (p *kcpPeer) Close() error {
	// wait after write finished
	p.lock.Lock()
	defer p.lock.Unlock()

	p.closed = true
	return p.c.Close() // close kcp conn
}

func (p *kcpPeer) GetId() int {
	return p.id
}

func (p *kcpPeer) Start() {
	p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond))
	for {
		buf := make([]byte, p.bufSize)
		n, err := p.c.Read(buf)
		if err != nil {
			if p.closed {
				// closed by peer.Close(), not need to remove peer from server
				break
			}
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
