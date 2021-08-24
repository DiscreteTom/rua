package websocket

import (
	"sync"

	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

type websocketPeer struct {
	id     int // peer id
	c      *websocket.Conn
	gs     rua.GameServer
	lock   sync.Mutex
	closed bool
	tag    string
	logger rua.Logger
}

func NewWebsocketPeer(c *websocket.Conn, gs rua.GameServer) *websocketPeer {
	return &websocketPeer{
		c:      c,
		gs:     gs,
		lock:   sync.Mutex{},
		closed: false,
		tag:    "websocket",
		logger: rua.GetDefaultLogger(),
	}
}

func (p *websocketPeer) WithLogger(l rua.Logger) *websocketPeer {
	p.logger = l
	return p
}

func (p *websocketPeer) WithTag(t string) *websocketPeer {
	p.tag = t
	return p
}

func (p *websocketPeer) SetTag(t string) {
	p.tag = t
}

func (p *websocketPeer) GetTag() string {
	return p.tag
}

func (p *websocketPeer) Activate(id int) {
	p.id = id
}

// Thread safe.
func (p *websocketPeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.c.WriteMessage(websocket.BinaryMessage, data)
}

func (p *websocketPeer) Close() error {
	// wait for write finished
	p.lock.Lock()
	defer p.lock.Unlock()

	p.closed = true
	return p.c.Close() // close websocket conn
}

func (p *websocketPeer) GetId() int {
	return p.id
}

func (p *websocketPeer) Start() {
	for {
		_, msg, err := p.c.ReadMessage()
		if err != nil {
			if !p.closed {
				// not closed by Close(), we should remove the peer
				p.logger.Error(err)
				if err := p.gs.RemovePeer(p.id); err != nil {
					p.logger.Error(err)
				}
			}
			break
		}

		p.gs.AppendPeerMsg(p.id, msg)
	}
}
