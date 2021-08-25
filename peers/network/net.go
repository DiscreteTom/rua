package network

import (
	"net"
	"sync"
	"time"

	"github.com/DiscreteTom/rua"
)

type netPeer struct {
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
func NewNetPeer(c net.Conn, gs rua.GameServer, bufSize int) *netPeer {
	return &netPeer{
		c:       c,
		gs:      gs,
		bufSize: bufSize,
		timeout: 0,
		lock:    sync.Mutex{},
		closed:  false,
		tag:     "net",
		logger:  rua.GetDefaultLogger(),
	}
}

func (p *netPeer) WithTimeout(ms int) *netPeer {
	p.timeout = ms
	return p
}

func (p *netPeer) WithTag(t string) *netPeer {
	p.tag = t
	return p
}

func (p *netPeer) WithLogger(l rua.Logger) *netPeer {
	p.logger = l
	return p
}

func (p *netPeer) SetTag(t string) {
	p.tag = t
}

func (p *netPeer) GetTag() string {
	return p.tag
}

func (p *netPeer) Activate(id int) {
	p.id = id
}

// Thread safe.
func (p *netPeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := p.c.Write(data)
	return err
}

func (p *netPeer) Close() error {
	// wait after write finished
	p.lock.Lock()
	defer p.lock.Unlock()

	p.closed = true
	return p.c.Close() // close connection
}

func (p *netPeer) GetId() int {
	return p.id
}

// Start the peer and wait.
func (p *netPeer) Start() {
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
