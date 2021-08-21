package websocket

import (
	"time"

	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

type websocketPeer struct {
	rc chan *rua.PeerMsg // receiver channel
	id int               // peer id
	c  *websocket.Conn
	gs rua.GameServer
}

func (p *websocketPeer) Activate(rc chan *rua.PeerMsg, id int) {
	p.rc = rc
	p.id = id
}

func (p *websocketPeer) Write(data []byte) error {
	return p.c.WriteMessage(websocket.BinaryMessage, data)
}

func (p *websocketPeer) Close() error {
	return p.c.Close() // close websocket conn
}

func (p *websocketPeer) Start() {
	for {
		_, msg, err := p.c.ReadMessage()
		if err != nil {
			p.gs.RemovePeer(p.id)
			break
		}

		p.rc <- &rua.PeerMsg{PeerId: p.id, Data: msg, Time: time.Now()}
	}
}
