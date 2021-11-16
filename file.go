package rua

import (
	"errors"
	"os"
)

type FileNode struct {
	handle   *Handle
	filename string
	stopRx   chan *StopPayload
	rx       chan *WritePayload
}

func NewFileNode(buffer uint) *FileNode {
	stopChan := make(chan *StopPayload)
	msgChan := make(chan *WritePayload, buffer)

	handle, _ := NewHandleBuilder().StopTx(stopChan).Tx(msgChan).Build()
	return &FileNode{
		handle:   handle,
		filename: "",
		stopRx:   stopChan,
		rx:       msgChan,
	}
}

func DefaultFileNode() *FileNode {
	return NewFileNode(16)
}

func (n *FileNode) Filename(name string) *FileNode {
	n.filename = name
	return n
}

func (n *FileNode) Handle() *Handle {
	return n.handle
}

func (n FileNode) Go() (*Handle, error) {
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
			case payload := <-rx:
				if _, err := file.Write(append(payload.Data, '\n')); err != nil {
					payload.Callback(err)
					loop = false
				} else if err = file.Sync(); err != nil {
					payload.Callback(err)
					loop = false
				} else {
					payload.Callback(nil)
				}
			case payload := <-stopRx:
				payload.Callback(nil)
				loop = false
			}
		}
	}()

	return n.handle, nil
}
