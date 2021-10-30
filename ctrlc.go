package rua

import (
	"os"
	"os/signal"
)

type Ctrlc struct {
	handler func()
}

func NewCtrlc() Ctrlc {
	return Ctrlc{handler: func() {}}
}

func (c Ctrlc) OnSignal(handler func()) Ctrlc {
	c.handler = handler
	return c
}

func (c Ctrlc) Go() {
	go func() {
		c.Wait()
	}()
}

func (self Ctrlc) Wait() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			self.handler()
			break
		}
	}()
}
