package rua

import (
	"log"
	"net"
	"sync"
	"time"
)

type basicPeer struct {
	id      int // peer id
	c       net.Conn
	gs      GameServer
	bufSize int
	timeout int // in ms
	lock    sync.Mutex
	closed  bool
}

func NewBasicPeer(c net.Conn, gs GameServer, bufSize int) *basicPeer {
	return &basicPeer{
		c:       c,
		gs:      gs,
		bufSize: bufSize,
		timeout: 0,
		lock:    sync.Mutex{},
		closed:  false,
	}
}

func (p *basicPeer) WithTimeout(ms int) *basicPeer {
	p.timeout = ms
	return p
}

func (p *basicPeer) Activate(id int) {
	p.id = id
}

// Thread safe.
func (p *basicPeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := p.c.Write(data)
	return err
}

func (p *basicPeer) Close() error {
	// wait after write finished
	p.lock.Lock()
	defer p.lock.Unlock()

	p.closed = true
	return p.c.Close() // close connection
}

func (p *basicPeer) GetId() int {
	return p.id
}

func (p *basicPeer) Start() {
	if p.timeout != 0 {
		p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond))
	}

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
		if p.timeout != 0 {
			p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond))
		}

		p.gs.AppendPeerMsg(p.id, buf[:n])
	}
}
