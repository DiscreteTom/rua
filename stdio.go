package rua

import (
	"bufio"
	"fmt"
	"os"
)

type StdioNode struct {
	msgHandler func([]byte)
	handle     WritableStoppableHandle
	msgChan    chan []byte
	stopChan   chan bool
}

func NewStdioNode(buffer uint, writeTimeoutMs int64) StdioNode {
	msgChan := make(chan []byte, buffer)
	stopChan := make(chan bool)

	return StdioNode{
		msgHandler: func(_ []byte) {},
		handle:     NewWritableStoppableHandle(msgChan, stopChan, writeTimeoutMs),
		msgChan:    msgChan,
		stopChan:   stopChan,
	}
}

func DefaultStdioNode() StdioNode {
	return NewStdioNode(16, 1000)
}

func (n StdioNode) OnMsg(f func([]byte)) StdioNode {
	n.msgHandler = f
	return n
}

func (n *StdioNode) Handle() WritableStoppableHandle {
	return n.handle
}

func (n StdioNode) Go() WritableStoppableHandle {
	stopChan := n.stopChan
	msgChan := n.msgChan

	// reader thread
	go func() {
		reader := bufio.NewReader(os.Stdin)
		loop := true
		for loop {
			line, err := reader.ReadString('\n')
			if len(line) == 0 || err != nil {
				break
			}
			select {
			case msgChan <- []byte(line[:len(line)-1]):
				continue
			case <-stopChan:
				loop = false
			}
		}
	}()

	// writer thread
	go func() {
		for data := range msgChan {
			fmt.Println(string(data))
		}
	}()

	return n.handle
}
