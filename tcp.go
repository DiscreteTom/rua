package rua

import (
	"bufio"
	"errors"
	"net"
)

type TcpListener struct {
	addr            string
	peerHandler     func(*TcpNode)
	peerWriteBuffer uint
	handle          *StopOnlyHandle
	stopRx          chan *StopPayload
}

func NewTcpListener(addr string) *TcpListener {
	stopChan := make(chan *StopPayload)
	handle, _ := NewHandleBuilder().StopTx(stopChan).BuildStopOnly()

	return &TcpListener{
		addr:            addr,
		peerHandler:     nil,
		peerWriteBuffer: 16,
		handle:          handle,
	}
}

func (l *TcpListener) PeerWriteBuffer(buffer uint) *TcpListener {
	l.peerWriteBuffer = buffer
	return l
}

func (l *TcpListener) OnNewPeer(f func(*TcpNode)) *TcpListener {
	l.peerHandler = f
	return l
}

func (l *TcpListener) Handle() *StopOnlyHandle {
	return l.handle
}

// Return error if missing `peerHandler`。
func (l *TcpListener) Go() (*StopOnlyHandle, error) {
	if l.peerHandler == nil {
		return nil, errors.New("missing peerHandler")
	}

	listener, err := net.Listen("tcp", l.addr)
	if err != nil {
		return nil, err
	}

	go func() {
		loop := true
		for loop {
			select {
			case payload := <-l.stopRx:
				payload.Callback(nil)
				loop = false
			default:
				conn, err := listener.Accept()
				if err != nil {
					loop = false
				} else {
					l.peerHandler(NewTcpNode(conn, l.peerWriteBuffer))
				}
			}
		}
	}()

	return l.handle, nil
}

type TcpNode struct {
	handle       *Handle
	conn         net.Conn
	inputHandler func([]byte)
	rx           chan *WritePayload
	stopRx       chan *StopPayload
}

func NewTcpNode(conn net.Conn, buffer uint) *TcpNode {
	msgChan := make(chan *WritePayload, buffer)
	stopChan := make(chan *StopPayload)

	handle, _ := NewHandleBuilder().Tx(msgChan).StopTx(stopChan).Build()
	return &TcpNode{
		handle:       handle,
		conn:         conn,
		inputHandler: func(b []byte) {},
		rx:           msgChan,
		stopRx:       stopChan,
	}
}

func (n *TcpNode) OnInput(f func([]byte)) *TcpNode {
	n.inputHandler = f
	return n
}

func (n *TcpNode) Handle() *Handle {
	return n.handle
}

func (n *TcpNode) Conn() net.Conn {
	return n.conn
}

func (n *TcpNode) Go() *Handle {
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
		reader := bufio.NewReader(n.conn)
		loop := true
		for loop {
			select {
			case <-readerStopper:
				loop = false
			default:
				line, err := reader.ReadString('\n')
				if len(line) == 0 || err != nil {
					break
				}
				line = line[:len(line)-1] // remove \n
				if line[len(line)-1] == '\r' {
					line = line[:len(line)-1] // remove \r
				}
				n.inputHandler([]byte(line))
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
				_, err := n.conn.Write(payload.Data)
				payload.Callback(err)
				if err != nil {
					loop = false
				}
			}
		}
	}()

	return n.handle
}
