package rua

import "sync"

type HandleIdManager struct {
	currentHandleId uint
}

func NewHandleIdManager() *HandleIdManager {
	return &HandleIdManager{currentHandleId: 0}
}

func (m *HandleIdManager) Next() uint {
	m.currentHandleId += 1
	return m.currentHandleId - 1
}

type Broadcaster struct {
	timeoutMs       uint64 // 0 means no timeout
	targets         map[uint]*Handle
	keepDeadTargets bool
	handleIdManager *HandleIdManager
	lock            *sync.Mutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		timeoutMs:       0,
		targets:         make(map[uint]*Handle),
		keepDeadTargets: false,
		handleIdManager: NewHandleIdManager(),
		lock:            &sync.Mutex{},
	}
}

func (b *Broadcaster) KeepDeadTargets(enable bool) *Broadcaster {
	b.keepDeadTargets = enable
	return b
}

func (b *Broadcaster) TimeoutMs(ms uint64) *Broadcaster {
	b.timeoutMs = ms
	return b
}

func (b *Broadcaster) AddTarget(handle *Handle) {
	b.AddTargetThen(handle, func(uint) {})
}

func (b *Broadcaster) AddTargetThen(handle *Handle, callback func(uint)) {
	go func() {
		b.lock.Lock()
		id := b.handleIdManager.Next()
		b.targets[id] = handle
		b.lock.Unlock()
		callback(id)
	}()
}

func (b *Broadcaster) RemoveTarget(id uint) {
	b.RemoveTargetThen(id, func(*Handle) {})
}

func (b *Broadcaster) RemoveTargetThen(id uint, callback func(*Handle)) {
	go func() {
		var target *Handle = nil
		ok := true
		b.lock.Lock()
		if target, ok = b.targets[id]; ok {
			delete(b.targets, id)
		}
		b.lock.Unlock()
		callback(target)
	}()
}

func (b *Broadcaster) Write(data []byte) {
	b.innerWrite(data, b.timeoutMs, func(error) {})
}

func (b *Broadcaster) WriteThen(data []byte, callback func(error)) {
	b.innerWrite(data, b.timeoutMs, callback)
}

func (b *Broadcaster) TimedWrite(data []byte, timeoutMs uint64) {
	b.innerWrite(data, timeoutMs, func(error) {})
}

func (b *Broadcaster) TimedWriteThen(data []byte, timeoutMs uint64, callback func(error)) {
	b.innerWrite(data, timeoutMs, callback)
}

func (b *Broadcaster) innerWrite(data []byte, timeoutMs uint64, callback func(error)) {
	go func() {
		b.lock.Lock()
		for id, target := range b.targets {
			_callback := func(err error) {
				if err != nil && !b.keepDeadTargets {
					go func() {
						b.RemoveTarget(id)
					}()
				}
				callback(err)
			}

			target.TimedWriteThen(data, timeoutMs, _callback)
		}
		b.lock.Unlock()
	}()

}
func (b *Broadcaster) StopAll() {
	b.StopAllThen(func(error) {})
}

func (b *Broadcaster) StopAllThen(callback func(error)) {
	go func() {
		b.lock.Lock()
		keys := make([]uint, 0, len(b.targets))
		for k := range b.targets {
			keys = append(keys, k)
		}

		for _, k := range keys {
			b.targets[k].StopThen(callback)
			delete(b.targets, k)
		}

		b.lock.Lock()
	}()
}
