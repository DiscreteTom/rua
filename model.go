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
	stopTx chan bool
}

func NewStoppableHandle(stopTx chan bool) StoppableHandle {
	return StoppableHandle{stopTx: stopTx}
}

func (h *StoppableHandle) Stop() {
	stopTx := h.stopTx
	go func() {
		stopTx <- true
	}()
}

type WritableStoppableHandle struct {
	StoppableHandle
	tx             chan []byte
	writeTimeoutMs int64
}

func NewWritableStoppableHandle(tx chan []byte, stopTx chan bool, writeTimeoutMs int64) WritableStoppableHandle {
	return WritableStoppableHandle{tx: tx, StoppableHandle: NewStoppableHandle(stopTx), writeTimeoutMs: writeTimeoutMs}
}

func (h *WritableStoppableHandle) Write(data []byte) error {
	c := make(chan bool)
	go Wait(h.writeTimeoutMs, c)
	select {
	case <-c:
		return errors.New("write time out")
	case h.tx <- data:
		return nil
	}
}

func Wait(ms int64, c chan<- bool) {
	time.Sleep(time.Millisecond * time.Duration(ms))
	c <- true
}
