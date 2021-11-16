package rua

import (
	"errors"
)

type WritePayload struct {
	data     []byte
	callback func(error)
}

func NewWritePayload(data []byte) *WritePayload {
	return &WritePayload{
		data:     data,
		callback: func(error) {},
	}
}

func (p *WritePayload) Callback(f func(error)) *WritePayload {
	p.callback = f
	return p
}

type StopPayload struct {
	callback func(error)
}

func NewStopPayload() *StopPayload {
	return &StopPayload{callback: func(error) {}}
}

func (p *StopPayload) Callback(f func(error)) *StopPayload {
	p.callback = f
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

func (b HandleBuilder) Build() (*Handle, error) {
	if b.stopTx == nil {
		return nil, errors.New("missing stopTx")
	}
	if b.tx == nil {
		return nil, errors.New("missing tx")
	}
	return &Handle{tx: b.tx, StopOnlyHandle: StopOnlyHandle{stopTx: b.stopTx}, timeoutMs: b.timeoutMs}, nil
}

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
		stopTx <- NewStopPayload().Callback(callback)
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
			case tx <- NewWritePayload(data).Callback(callback):
			case <-c:
				callback(errors.New("write timeout"))
			}
		} else {
			// no timeout
			tx <- NewWritePayload(data).Callback(callback)
		}
	}()
}
