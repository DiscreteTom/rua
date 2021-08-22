package websocket

import (
	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

type websocketPeer struct {
	id int // peer id
	c  *websocket.Conn
	gs rua.GameServer
}

func (p *websocketPeer) Activate(id int) {
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

		p.gs.AppendPeerMsg(p.id, msg)
	}
}
