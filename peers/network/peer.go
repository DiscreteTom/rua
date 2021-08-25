package rua

import (
	"net"
	"sync"
	"time"

	"github.com/DiscreteTom/rua"
)

type basicPeer struct {
	id      int // peer id
	c       net.Conn
	gs      rua.GameServer
	bufSize int
	timeout int // in ms
	lock    sync.Mutex
	closed  bool
	tag     string
	logger  rua.Logger
}

// Create a peer with a connection of `net.Conn`.
func NewBasicPeer(c net.Conn, gs rua.GameServer, bufSize int) *basicPeer {
	return &basicPeer{
		c:       c,
		gs:      gs,
		bufSize: bufSize,
		timeout: 0,
		lock:    sync.Mutex{},
		closed:  false,
		tag:     "basic",
		logger:  rua.GetDefaultLogger(),
	}
}

func (p *basicPeer) WithTimeout(ms int) *basicPeer {
	p.timeout = ms
	return p
}

func (p *basicPeer) WithTag(t string) *basicPeer {
	p.tag = t
	return p
}

func (p *basicPeer) WithLogger(l rua.Logger) *basicPeer {
	p.logger = l
	return p
}

func (p *basicPeer) SetTag(t string) {
	p.tag = t
}

func (p *basicPeer) GetTag() string {
	return p.tag
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

// Start the peer and wait.
func (p *basicPeer) Start() {
	if p.timeout != 0 {
		if err := p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond)); err != nil {
			p.logger.Error(err)
		}
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
				p.logger.Infof("peer[%d] timeout", p.id)
			}
			if err := p.gs.RemovePeer(p.id); err != nil {
				p.logger.Error(err)
			}
			break
		}
		if p.timeout != 0 {
			if err := p.c.SetReadDeadline(time.Now().Add(time.Duration(p.timeout) * time.Millisecond)); err != nil {
				p.logger.Error(err)
			}
		}

		p.gs.AppendPeerMsg(p.id, buf[:n])
	}
}
