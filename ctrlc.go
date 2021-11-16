package rua

import (
	"os"
	"os/signal"
)

type Ctrlc struct {
	handle        *StopOnlyHandle
	stopRx        chan *StopPayload
	signalHandler func()
}

func NewCtrlc() *Ctrlc {
	stopChan := make(chan *StopPayload)
	handle, _ := NewHandleBuilder().StopTx(stopChan).BuildStopOnly()
	return &Ctrlc{signalHandler: func() {}, stopRx: stopChan, handle: handle}
}

func (c *Ctrlc) OnSignal(handler func()) *Ctrlc {
	c.signalHandler = handler
	return c
}

func (c *Ctrlc) Handle() *StopOnlyHandle {
	return c.handle
}

func (c Ctrlc) Go() *StopOnlyHandle {
	go func() {
		c.Wait()
	}()
	return c.handle
}

func (c Ctrlc) Wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	for range ch {
		c.signalHandler()
		break
	}
}
