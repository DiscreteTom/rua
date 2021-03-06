package rua

import (
	"errors"
)

type WritePayload struct {
	Data     []byte
	Callback func(error)
}

func NewWritePayload(data []byte) *WritePayload {
	return &WritePayload{
		Data:     data,
		Callback: func(error) {},
	}
}

func (p *WritePayload) WithCallback(f func(error)) *WritePayload {
	p.Callback = f
	return p
}

type StopPayload struct {
	Callback func(error)
}

func NewStopPayload() *StopPayload {
	return &StopPayload{Callback: func(error) {}}
}

func (p *StopPayload) WithCallback(f func(error)) *StopPayload {
	p.Callback = f
	return p
}

type HandleBuilder struct {
	tx        chan *WritePayload
	stopTx    chan *StopPayload
	timeoutMs uint64 // 0 means no timeout
}

func NewHandleBuilder() *HandleBuilder {
	return &HandleBuilder{
		tx:        nil,
		stopTx:    nil,
		timeoutMs: 0,
	}
}

func (b *HandleBuilder) Tx(tx chan *WritePayload) *HandleBuilder {
	b.tx = tx
	return b
}

func (b *HandleBuilder) StopTx(stopTx chan *StopPayload) *HandleBuilder {
	b.stopTx = stopTx
	return b
}
func (b *HandleBuilder) TimeoutMs(timeoutMs uint64) *HandleBuilder {
	b.timeoutMs = timeoutMs
	return b
}

// Return error if missing `stopTx` or `tx`.
func (b HandleBuilder) Build() (*Handle, error) {
	if b.stopTx == nil {
		return nil, errors.New("missing stopTx")
	}
	if b.tx == nil {
		return nil, errors.New("missing tx")
	}
	return &Handle{tx: b.tx, StopOnlyHandle: StopOnlyHandle{stopTx: b.stopTx}, timeoutMs: b.timeoutMs}, nil
}

// Return error if missing `stopTx`.
func (b HandleBuilder) BuildStopOnly() (*StopOnlyHandle, error) {
	if b.stopTx == nil {
		return nil, errors.New("missing stopTx")
	}
	return &StopOnlyHandle{stopTx: b.stopTx}, nil
}

type StopOnlyHandle struct {
	stopTx chan *StopPayload
}

func (h StopOnlyHandle) Stop() {
	stopTx := h.stopTx
	go func() {
		stopTx <- NewStopPayload()
	}()
}

func (h StopOnlyHandle) StopThen(callback func(error)) {
	stopTx := h.stopTx
	go func() {
		stopTx <- NewStopPayload().WithCallback(callback)
	}()
}

type Handle struct {
	StopOnlyHandle
	tx        chan *WritePayload
	timeoutMs uint64
}

func (h *Handle) SetTimeoutMs(ms uint64) {
	h.timeoutMs = ms
}

func (h *Handle) ClearTimeout() {
	h.timeoutMs = 0
}

func (h *Handle) Write(data []byte) {
	innerWrite(h.tx, data, h.timeoutMs, func(error) {})
}

func (h *Handle) WriteThen(data []byte, callback func(error)) {
	innerWrite(h.tx, data, h.timeoutMs, callback)
}

func (h *Handle) TimedWrite(data []byte, timeoutMs uint64) {
	innerWrite(h.tx, data, timeoutMs, func(error) {})
}

func (h *Handle) TimedWriteThen(data []byte, timeoutMs uint64, callback func(error)) {
	innerWrite(h.tx, data, timeoutMs, callback)
}

func innerWrite(tx chan *WritePayload, data []byte, timeoutMs uint64, callback func(error)) {
	go func() {
		if timeoutMs != 0 {
			c := make(chan bool)
			go Wait(timeoutMs, c)
			select {
			case tx <- NewWritePayload(data).WithCallback(callback):
			case <-c:
				callback(errors.New("write timeout"))
			}
		} else {
			// no timeout
			tx <- NewWritePayload(data).WithCallback(callback)
		}
	}()
}
