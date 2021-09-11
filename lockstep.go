package rua

import (
	"sync"
	"time"
)

type LockstepServer struct {
	*EventDrivenServer
	peerMsgs      []PeerMsg   // msgs from peers in one step
	peerMsgsLock  *sync.Mutex // prevent concurrent access
	currentStep   int         // current step number, start from 0
	stepLength    int         // how many ms to wait after a step
	maxStepLength int
	minStepLength int
	onStepHandler func(peerMsgs []PeerMsg) // lifecycle hook
}

func NewLockstepServer() *LockstepServer {
	return &LockstepServer{
		EventDrivenServer: NewEventDrivenServer().WithName("LockstepServer"),
		peerMsgs:          []PeerMsg{},
		peerMsgsLock:      &sync.Mutex{},
		currentStep:       0,
		stepLength:        100, // 10 step/second
		maxStepLength:     200, // 5 step/second
		minStepLength:     8,   // ~120 step/second
		onStepHandler:     func(peerMsgs []PeerMsg) {},
	}
}

// Set the current step length.
// The step length won't be higher than `maxStepLength` and lower than `minStepLength`.
func (s *LockstepServer) WithStepLength(stepLength int) *LockstepServer {
	if stepLength > s.maxStepLength {
		s.stepLength = s.maxStepLength
	} else if stepLength < s.minStepLength {
		s.stepLength = s.minStepLength
	} else {
		s.stepLength = stepLength
	}
	return s
}

// Set the max step length and ensure the current step length is valid.
func (s *LockstepServer) WithMaxStepLength(maxStepLength int) *LockstepServer {
	s.maxStepLength = maxStepLength
	if s.stepLength > s.maxStepLength {
		s.stepLength = s.maxStepLength
	}
	return s
}

// Set the min step length and ensure the current step length is valid.
func (s *LockstepServer) WithMinStepLength(minStepLength int) *LockstepServer {
	s.minStepLength = minStepLength
	if s.stepLength < s.minStepLength {
		s.stepLength = s.minStepLength
	}
	return s
}

func (s *LockstepServer) CurrentStepLength() int {
	return s.stepLength
}

func (s *LockstepServer) CurrentStep() int {
	return s.currentStep
}

// Register lifecycle hook.
// This hook may be triggered concurrently.
func (s *LockstepServer) OnStep(f func(peerMsgs []PeerMsg)) *LockstepServer {
	s.onStepHandler = f
	return s
}

// Return errors from peer.Close() when stop the server.
func (s *LockstepServer) Start() (errs []error) {
	errs = []error{}

	timer := time.NewTimer(time.Duration(s.stepLength))

	s.logger.Infof("%s started, step length: %dms", s.name, s.stepLength)

	loop := true
	for loop {
		select {
		case <-timer.C:
			// retrieve peer msgs and reset
			s.peerMsgsLock.Lock()
			currentPeerMsgs := s.peerMsgs
			s.peerMsgs = []PeerMsg{}
			s.peerMsgsLock.Unlock()

			// handle lifecycle hook
			s.onStepHandler(currentPeerMsgs)

			s.currentStep++
			// reset timer
			timer = time.NewTimer(time.Duration(s.stepLength) * time.Millisecond)
		case <-s.stop:
			loop = false
		}
	}

	// close all peers
	for id, peer := range s.peers {
		if err := peer.Close(); err != nil {
			errs = append(errs, err)
		}
		delete(s.peers, id)
	}
	return errs
}

func (s *LockstepServer) Stop() {
	s.stop <- true
}

func (s *LockstepServer) AppendPeerMsg(p Peer, d []byte) {
	peerMsg := PeerMsg{Peer: p, Data: d, Time: time.Now()}

	// handle lifecycle hook
	// this hook can modify peerMsg before append
	s.beforeProcPeerMsgHandler(&peerMsg)

	s.peerMsgsLock.Lock()
	s.peerMsgs = append(s.peerMsgs, peerMsg)
	s.peerMsgsLock.Unlock()

	// handle lifecycle hook
	s.onPeerMsgHandler(&peerMsg)
}
