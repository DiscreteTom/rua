package rua

import (
	"errors"
	"time"
)

type Ticker struct {
	tickHandler func(uint64)
	intervalMs  uint64
	stopRx      chan *StopPayload
	handle      *StopOnlyHandle
}

func NewTicker(intervalMs uint64) *Ticker {
	stopChan := make(chan *StopPayload)

	handle, _ := NewHandleBuilder().StopTx(stopChan).BuildStopOnly()
	return &Ticker{
		tickHandler: nil,
		intervalMs:  intervalMs,
		stopRx:      stopChan,
		handle:      handle,
	}
}

func DefaultTicker() *Ticker {
	return NewTicker(1000)
}

func (t *Ticker) IntervalMs(ms uint64) *Ticker {
	t.intervalMs = ms
	return t
}

func (t *Ticker) Handle() *StopOnlyHandle {
	return t.handle
}

func (t *Ticker) OnTick(f func(uint64)) *Ticker {
	t.tickHandler = f
	return t
}

// Return error if missing `tickHandler`.
func (t *Ticker) Go() (*StopOnlyHandle, error) {
	if t.tickHandler == nil {
		return nil, errors.New("missing tickHandler")
	}

	go func() {
		var current uint64 = 0
		ticker := time.NewTicker(time.Duration(t.intervalMs))
		loop := true
		for loop {
			select {
			case <-ticker.C:
				t.tickHandler(current)
				current += 1
			case payload := <-t.stopRx:
				payload.Callback(nil)
				loop = false
			}
		}
	}()

	return t.handle, nil
}
