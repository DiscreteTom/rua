package rua

import (
	"errors"
	"time"
)

type Writable interface {
	Write(data []byte) error
}

type Stoppable interface {
	Stop()
}

type StoppableHandle struct {
	stopChan chan bool
}

func NewStoppableHandle(stopChan chan bool) StoppableHandle {
	return StoppableHandle{stopChan: stopChan}
}

func (h *StoppableHandle) Stop() {
	stopChan := h.stopChan
	go func() {
		stopChan <- true
	}()
}

type WritableStoppableHandle struct {
	StoppableHandle
	msgChan        chan []byte
	writeTimeoutMs int64
}

func NewWritableStoppableHandle(msgChan chan []byte, stopChan chan bool, writeTimeoutMs int64) WritableStoppableHandle {
	return WritableStoppableHandle{msgChan: msgChan, StoppableHandle: NewStoppableHandle(stopChan), writeTimeoutMs: writeTimeoutMs}
}

func (h *WritableStoppableHandle) Write(data []byte) error {
	c := make(chan bool)
	go Wait(h.writeTimeoutMs, c)
	select {
	case <-c:
		return errors.New("write time out")
	case h.msgChan <- data:
		return nil
	}
}

func Wait(ms int64, c chan<- bool) {
	time.Sleep(time.Millisecond * time.Duration(ms))
	c <- true
}
