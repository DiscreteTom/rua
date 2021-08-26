package websocket

import (
	"sync"

	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

func NewWebsocketPeer(c *websocket.Conn, gs rua.GameServer) *rua.BasicPeer {
	lock := sync.Mutex{}
	closed := false

	return rua.NewBasicPeer(gs).
		WithTag("websocket").
		OnWrite(func(data []byte, p *rua.BasicPeer) error {
			// prevent concurrent write
			lock.Lock()
			defer lock.Unlock()

			return c.WriteMessage(websocket.BinaryMessage, data)
		}).
		OnClose(func(p *rua.BasicPeer) error {
			// wait for write finished
			lock.Lock()
			defer lock.Unlock()

			closed = true
			return c.Close() // close websocket conn
		}).
		OnStart(func(p *rua.BasicPeer) {
			for {
				_, msg, err := c.ReadMessage()
				if err != nil {
					if !closed {
						// not closed by Close(), we should remove the peer
						p.GetLogger().Error(err)
						if err := p.GetGameServer().RemovePeer(p.GetId()); err != nil {
							p.GetLogger().Error(err)
						}
					}
					break
				}

				p.GetGameServer().AppendPeerMsg(p.GetId(), msg)
			}
		})
}
