package websocket

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"

	"github.com/gorilla/websocket"
)

type WebsocketPeer struct {
	*peer.SafePeer
	closed bool
	c      *websocket.Conn
}

func NewWebsocketPeer(c *websocket.Conn, gs rua.GameServer) *WebsocketPeer {
	wp := &WebsocketPeer{
		SafePeer: peer.NewSafePeer(gs),
		closed:   false,
		c:        c,
	}

	wp.SafePeer.
		OnWriteSafe(func(data []byte) error {
			return wp.c.WriteMessage(websocket.BinaryMessage, data)
		}).
		OnCloseSafe(func() error {
			wp.closed = true
			return wp.c.Close() // close websocket conn
		}).
		OnStart(func() {
			for {
				_, msg, err := wp.c.ReadMessage()
				if err != nil {
					// normally closed by server or client?
					if !websocket.IsCloseError(err, websocket.CloseNoStatusReceived) {
						wp.Logger().Error(err)
					}
					if !wp.closed {
						// not closed by Close(), we should remove the peer
						if err := wp.GameServer().RemovePeer(wp.Id()); err != nil {
							wp.Logger().Error(err)
						}
					}
					break
				}

				wp.GameServer().AppendPeerMsg(wp, msg)
			}
		}).
		WithTag("websocket")

	return wp
}
