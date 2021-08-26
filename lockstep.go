package rua

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

type LockstepServer struct {
	stop                     chan bool
	handleKeyboardInterrupt  bool
	peers                    map[int]Peer // peer id starts from 0
	peerLock                 sync.Mutex   // prevent concurrent access
	peerMsgs                 []PeerMsg    // msgs from peers
	peerMsgsLock             sync.Mutex   // prevent concurrent access
	currentStep              int          // current step number, start from 0
	stepLength               int          // how many ms to wait after a step
	maxStepLength            int
	minStepLength            int
	beforeAddPeerHandler     func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	afterAddPeerHandler      func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	beforeRemovePeerHandler  func(step int, targetId int, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	afterRemovePeerHandler   func(step int, targetId int, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	beforeProcPeerMsgHandler func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer)         // lifecycle hook
	onPeerMsgHandler         func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer)         // lifecycle hook
	onStepHandler            func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer) // lifecycle hook
	logger                   Logger
}

func NewLockStepServer() *LockstepServer {
	return &LockstepServer{
		stop:                     make(chan bool),
		handleKeyboardInterrupt:  false,
		peers:                    map[int]Peer{},
		peerLock:                 sync.Mutex{},
		peerMsgs:                 []PeerMsg{},
		peerMsgsLock:             sync.Mutex{},
		currentStep:              0,
		stepLength:               33,  // ~30 step/second
		maxStepLength:            100, // ~10 step/second
		minStepLength:            8,   // ~120 step/second
		beforeAddPeerHandler:     func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer) {},
		afterAddPeerHandler:      func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer) {},
		beforeRemovePeerHandler:  func(step int, targetId int, peers map[int]Peer, s *LockstepServer) {},
		afterRemovePeerHandler:   func(step int, targetId int, peers map[int]Peer, s *LockstepServer) {},
		beforeProcPeerMsgHandler: func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer) {},
		onPeerMsgHandler:         func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer) {},
		onStepHandler:            func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer) {},
		logger:                   GetDefaultLogger(),
	}
}

func (s *LockstepServer) WithLogger(l Logger) *LockstepServer {
	s.logger = l
	return s
}

func (s *LockstepServer) GetLogger() Logger {
	return s.logger
}

// Set the current step length.
// The step length won't be higher than `maxStepLength` and lower than `minStepLength`.
func (s *LockstepServer) SetStepLength(stepLength int) *LockstepServer {
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
func (s *LockstepServer) SetMaxStepLength(maxStepLength int) *LockstepServer {
	s.maxStepLength = maxStepLength
	if s.stepLength > s.maxStepLength {
		s.stepLength = s.maxStepLength
	}
	return s
}

// Set the min step length and ensure the current step length is valid.
func (s *LockstepServer) SetMinStepLength(minStepLength int) *LockstepServer {
	s.minStepLength = minStepLength
	if s.stepLength < s.minStepLength {
		s.stepLength = s.minStepLength
	}
	return s
}

func (s *LockstepServer) SetHandleKeyboardInterrupt(enable bool) *LockstepServer {
	s.handleKeyboardInterrupt = enable
	return s
}

func (s *LockstepServer) GetCurrentStepLength() int {
	return s.stepLength
}

// Activate a peer, allocate a peerId and manage the peer's lifecycle.
func (s *LockstepServer) AddPeer(p Peer) {
	s.peerLock.Lock()

	// allocate a peerId
	peerId := 0
	for {
		_, ok := s.peers[peerId]
		if !ok {
			break
		}
		peerId++
	}

	p.SetId(peerId)
	s.beforeAddPeerHandler(s.currentStep, p, s.peers, s)
	s.peers[peerId] = p

	s.peerLock.Unlock()
	go p.Start()

	s.afterAddPeerHandler(s.currentStep, p, s.peers, s)
}

// Close the peer and untrack it. Return err if peer not exist.
func (s *LockstepServer) RemovePeer(peerId int) (err error) {
	s.beforeRemovePeerHandler(s.currentStep, peerId, s.peers, s)

	s.peerLock.Lock()
	if peer, ok := s.peers[peerId]; ok {
		if err := peer.Close(); err != nil {
			s.logger.Error(err)
		}
		delete(s.peers, peerId)
	} else {
		err = errors.New("peer not exist")
	}
	s.peerLock.Unlock()

	s.afterRemovePeerHandler(s.currentStep, peerId, s.peers, s)
	return
}

func (s *LockstepServer) GetPeerCount() int {
	return len(s.peers)
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, but `peers` not contains the new peer.
// This hook won't be triggered concurrently.
func (s *LockstepServer) BeforeAddPeer(f func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer)) *LockstepServer {
	s.beforeAddPeerHandler = f
	return s
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, and `peers` contains the new peer.
// This hook may be triggered concurrently.
func (s *LockstepServer) AfterAddPeer(f func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer)) *LockstepServer {
	s.afterAddPeerHandler = f
	return s
}

// Register lifecycle hook.
// The target peer may has been closed.
// The target peer may not exist.
// This hook may be triggered concurrently.
func (s *LockstepServer) BeforeRemovePeer(f func(step int, targetId int, peers map[int]Peer, s *LockstepServer)) *LockstepServer {
	s.beforeRemovePeerHandler = f
	return s
}

// Register lifecycle hook.
// The target peer may not exist.
// If it exists, it must been closed, and been removed from `peers`.
// This hook may be triggered concurrently.
func (s *LockstepServer) AfterRemovePeer(f func(step int, targetId int, peers map[int]Peer, s *LockstepServer)) *LockstepServer {
	s.afterRemovePeerHandler = f
	return s
}

// Register lifecycle hook.
// You can modify or enrich the peer message before process it.
// This hook may be triggered concurrently.
func (s *LockstepServer) BeforeProcPeerMsg(f func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer)) *LockstepServer {
	s.beforeProcPeerMsgHandler = f
	return s
}

// Register lifecycle hook.
// This hook may be triggered concurrently.
func (s *LockstepServer) OnPeerMsg(f func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer)) *LockstepServer {
	s.onPeerMsgHandler = f
	return s
}

// Register lifecycle hook.
// This hook may be triggered concurrently.
func (s *LockstepServer) OnStep(f func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer)) *LockstepServer {
	s.onStepHandler = f
	return s
}

// Return errors from peer.Close() when stop the server.
func (s *LockstepServer) Start() (errs []error) {
	errs = []error{}

	timer := time.NewTimer(time.Duration(s.stepLength))

	// keyboard interrupt handler channel
	kbc := make(chan os.Signal, 1)
	signal.Notify(kbc, os.Interrupt)

	s.logger.Info("lockstep server started, step length:", s.stepLength, "ms")

	loop := true
	for loop {
		select {
		case <-timer.C:
			// handle lifecycle hook
			s.onStepHandler(s.currentStep, s.peers, s.peerMsgs, s)

			s.currentStep++
			// reset msgs
			s.peerMsgs = []PeerMsg{}
			// reset timer
			timer = time.NewTimer(time.Duration(s.stepLength) * time.Millisecond)
		case <-kbc:
			if s.handleKeyboardInterrupt {
				loop = false
			}
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

func (s *LockstepServer) AppendPeerMsg(peerId int, d []byte) {
	peerMsg := PeerMsg{PeerId: peerId, Data: d, Time: time.Now()}

	// handle lifecycle hook
	// this hook can modify peerMsg before append
	s.beforeProcPeerMsgHandler(s.currentStep, s.peers, &peerMsg, s)

	s.peerMsgsLock.Lock()
	s.peerMsgs = append(s.peerMsgs, peerMsg)
	s.peerMsgsLock.Unlock()

	// handle lifecycle hook
	s.onPeerMsgHandler(s.currentStep, s.peers, &peerMsg, s)
}
