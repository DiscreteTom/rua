package rua

import "time"

type Lockstep struct {
	stepHandler  func(uint64)
	stepLengthMs uint64
	stopRx       chan bool
	handle       StoppableHandle
}

func NewLockstep() Lockstep {
	stopChan := make(chan bool)
	return Lockstep{
		stepHandler:  func(_ uint64) {},
		stepLengthMs: 1000,
		stopRx:       stopChan,
		handle:       NewStoppableHandle(stopChan),
	}
}

func (l Lockstep) StepLengthMs(ms uint64) Lockstep {
	l.stepLengthMs = ms
	return l
}

func (l Lockstep) OnStep(f func(uint64)) Lockstep {
	l.stepHandler = f
	return l
}

func (l Lockstep) Go() StoppableHandle {
	stepHandler := l.stepHandler
	stepLengthMs := l.stepLengthMs
	stopRx := l.stopRx

	go func() {
		var current uint64 = 0
		loop := true
		timer := time.NewTimer(time.Duration(stepLengthMs) * time.Millisecond)

		for loop {
			select {
			case <-timer.C:
				stepHandler(current)
				current += 1
				timer = time.NewTimer(time.Duration(stepLengthMs) * time.Millisecond)
			case <-stopRx:
				loop = false
			}
		}
	}()

	return l.handle
}
