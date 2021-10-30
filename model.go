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
	stop_tx chan bool
}

func NewStoppableHandle(stop_tx chan bool) StoppableHandle {
	return StoppableHandle{stop_tx: stop_tx}
}

func (h *StoppableHandle) Stop() {
	stop_tx := h.stop_tx
	go func() {
		stop_tx <- true
	}()
}

type WritableStoppableHandle struct {
	StoppableHandle
	tx             chan []byte
	writeTimeoutMs int64
}

func NewWritableStoppableHandle(tx chan []byte, stop_tx chan bool, writeTimeoutMs int64) WritableStoppableHandle {
	return WritableStoppableHandle{tx: tx, StoppableHandle: NewStoppableHandle(stop_tx), writeTimeoutMs: writeTimeoutMs}
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
