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
}

func (p *websocketPeer) Activate(id int) {
	p.id = id
	p.lock = sync.Mutex{}
	p.closed = false
}

func (p *websocketPeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.c.WriteMessage(websocket.BinaryMessage, data)
}

func (p *websocketPeer) Close() error {
	p.closed = true
	return p.c.Close() // close websocket conn
}

func (p *websocketPeer) GetId() int {
	return p.id
}

func (p *websocketPeer) Start() {
	for {
		_, msg, err := p.c.ReadMessage()
		if err != nil && !p.closed {
			// not closed by Close(), we should remove the peer
			p.gs.RemovePeer(p.id)
			break
		}

		p.gs.AppendPeerMsg(p.id, msg)
	}
}
