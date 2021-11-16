package rua

import (
	"bufio"
	"io"
	"os"
	"time"
)

type TailNode struct {
	handle          *StopOnlyHandle
	filename        string
	stopRx          chan *StopPayload
	lineHandler     func([]byte)
	checkIntervalMs uint64
}

func NewTailNode(filename string) *TailNode {
	stopChan := make(chan *StopPayload)
	handle, _ := NewHandleBuilder().StopTx(stopChan).BuildStopOnly()
	return &TailNode{
		handle:          handle,
		filename:        filename,
		lineHandler:     nil,
		stopRx:          stopChan,
		checkIntervalMs: 10,
	}
}

func (n *TailNode) OnNewLine(f func([]byte)) *TailNode {
	n.lineHandler = f
	return n
}

func (n *TailNode) CheckIntervalMs(ms uint64) *TailNode {
	n.checkIntervalMs = ms
	return n
}

func (n *TailNode) Handle() *StopOnlyHandle {
	return n.handle
}

func (n *TailNode) Go() (*StopOnlyHandle, error) {
	file, err := os.OpenFile(n.filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	if _, err = file.Seek(0, 2); err != nil {
		file.Close()
		return nil, err
	}

	go func() {
		loop := true
		reader := bufio.NewReader(file)
		for loop {
			select {
			case payload := <-n.stopRx:
				payload.Callback(nil)
				loop = false
			default:
				line, err := reader.ReadString('\n')
				if len(line) == 0 && err == io.EOF {
					time.Sleep(time.Millisecond * time.Duration(n.checkIntervalMs))
				} else if len(line) == 0 || err != nil {
					loop = false
				} else {
					line = line[:len(line)-1] // remove \n
					if line[len(line)-1] == '\r' {
						line = line[:len(line)-1] // remove \r
					}
					n.lineHandler([]byte(line))
				}
			}
		}
	}()

	return n.handle, nil
}
