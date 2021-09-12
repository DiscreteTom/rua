package websocket

import (
	"errors"

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
			if wp.closed {
				return errors.New("peer already closed")
			}
			return wp.c.WriteMessage(websocket.BinaryMessage, data)
		}).
		OnCloseSafe(func() error {
			if wp.closed {
				return errors.New("peer already closed")
			}
			wp.closed = true
			return wp.c.Close() // close websocket conn
		}).
		OnStart(func() {
			for {
				_, msg, err := wp.c.ReadMessage()
				if err != nil {
					if !wp.closed { // not closed by Close()
						// normally closed by client?
						if !websocket.IsCloseError(err, websocket.CloseNoStatusReceived) {
							wp.Logger().Error("rua.WebsocketPeer.OnStart:", err)
						} else {
							wp.Logger().Info("rua.WebsocketPeer: peer", wp.Id(), "disconnected")
						}
						// we should remove the peer
						if err := wp.GameServer().RemovePeer(wp.Id()); err != nil {
							wp.Logger().Error("rua.WebsocketPeer.OnStart.RemovePeer:", err)
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
