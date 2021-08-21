package websocket

import (
	"DiscreteTom/rua/pkg/model"
	"time"

	"github.com/gorilla/websocket"
)

type websocketPeer struct {
	rc chan *model.PeerMsg // receiver channel
	id int                 // peer id
	c  *websocket.Conn
	gs model.GameServer
}

func (p *websocketPeer) Activate(rc chan *model.PeerMsg, id int) {
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

		p.rc <- &model.PeerMsg{PeerId: p.id, Data: msg, Time: time.Now()}
	}
}
