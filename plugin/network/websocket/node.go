package websocket

import (
	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

type WsNode struct {
	handle     *rua.Handle
	c          *websocket.Conn
	rx         chan *rua.WritePayload
	stopRx     chan *rua.StopPayload
	msgHandler func([]byte)
}

func NewWsNode(c *websocket.Conn, buffer uint) *WsNode {
	msgChan := make(chan *rua.WritePayload, buffer)
	stopChan := make(chan *rua.StopPayload)
	handle, _ := rua.NewHandleBuilder().Tx(msgChan).StopTx(stopChan).Build()

	return &WsNode{
		c:          c,
		handle:     handle,
		rx:         msgChan,
		stopRx:     stopChan,
		msgHandler: func(b []byte) {},
	}
}

func (n *WsNode) OnMsg(f func([]byte)) *WsNode {
	n.msgHandler = f
	return n
}

func (n *WsNode) Handle() *rua.Handle {
	return n.handle
}

func (n *WsNode) Go() *rua.Handle {
	readerStopper := make(chan bool)
	writerStopper := make(chan bool)

	// stopper thread
	go func() {
		payload := <-n.stopRx
		readerStopper <- true
		writerStopper <- true
		payload.Callback(nil)
	}()

	// reader thread
	go func() {
		loop := true
		for loop {
			select {
			case <-readerStopper:
				loop = false
			default:
				_, msg, err := n.c.ReadMessage()
				if len(msg) == 0 || err != nil {
					break
				}
				n.msgHandler([]byte(msg))
			}
		}
		writerStopper <- true
	}()

	// writer thread
	go func() {
		loop := true
		for loop {
			select {
			case <-writerStopper:
				loop = false
			case payload := <-n.rx:
				err := n.c.WriteMessage(websocket.BinaryMessage, payload.Data)
				payload.Callback(err)
				if err != nil {
					loop = false
				}
			}
		}
	}()

	return n.handle
}
