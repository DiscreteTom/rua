package rua

import (
	"errors"
	"os"
)

type FileNode struct {
	handle   WritableStoppableHandle
	filename string
	stopRx   chan bool
	rx       chan []byte
}

func NewFileNode(buffer uint, timeout int64) FileNode {
	stop_chan := make(chan bool)
	msg_chan := make(chan []byte, buffer)
	return FileNode{
		handle:   NewWritableStoppableHandle(msg_chan, stop_chan, timeout),
		filename: "",
		stopRx:   stop_chan,
		rx:       msg_chan,
	}
}

func DefaultFileNode() FileNode {
	return NewFileNode(16, 1000)
}

func (n FileNode) Filename(name string) FileNode {
	n.filename = name
	return n
}

func (n *FileNode) Handle() WritableStoppableHandle {
	return n.handle
}

func (n FileNode) Go() (*WritableStoppableHandle, error) {
	if len(n.filename) == 0 {
		return nil, errors.New("missing filename")
	}

	file, err := os.OpenFile(n.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	rx := n.rx
	stopRx := n.stopRx

	go func() {
		loop := true
		for loop {
			select {
			case data := <-rx:
				if _, err := file.Write(data); err != nil {
					loop = false
				} else if err = file.Sync(); err != nil {
					loop = false
				}
			case <-stopRx:
				loop = false
			}
		}
	}()

	return &n.handle, nil
}
