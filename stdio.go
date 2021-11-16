package rua

import (
	"bufio"
	"fmt"
	"os"
)

type StdioNode struct {
	inputHandler func([]byte)
	handle       *Handle
	rx           chan *WritePayload
	stopRx       chan *StopPayload
}

func NewStdioNode(buffer uint) *StdioNode {
	msgChan := make(chan *WritePayload, buffer)
	stopChan := make(chan *StopPayload)

	handle, _ := NewHandleBuilder().StopTx(stopChan).Tx(msgChan).Build()

	return &StdioNode{
		inputHandler: nil,
		handle:       handle,
		rx:           msgChan,
		stopRx:       stopChan,
	}
}

func DefaultStdioNode() *StdioNode {
	return NewStdioNode(16)
}

func (n *StdioNode) OnInput(f func([]byte)) *StdioNode {
	n.inputHandler = f
	return n
}

func (n *StdioNode) Handle() *Handle {
	return n.handle
}

func (n StdioNode) Go() *Handle {
	readerStopper := make(chan bool)
	writerStopper := make(chan bool)

	stopRx := n.stopRx
	rx := n.rx
	inputHandler := n.inputHandler

	// stopper thread
	go func() {
		payload := <-stopRx
		readerStopper <- true
		writerStopper <- true
		payload.Callback(nil)
	}()

	// reader thread
	if inputHandler != nil {
		go func() {
			reader := bufio.NewReader(os.Stdin)
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
					inputHandler([]byte(line))
				}
			}
		}()
	}

	// writer thread
	go func() {
		loop := true
		for loop {
			select {
			case <-writerStopper:
				loop = false
			case payload := <-rx:
				fmt.Println(string(payload.Data))
			}
		}
	}()

	return n.handle
}
