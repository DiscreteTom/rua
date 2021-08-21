package rua

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"time"
)

type LockstepServer struct {
	peers                   map[int]Peer
	stop                    chan bool
	handleKeyboardInterrupt bool
	rc                      chan *PeerMsg // receiver channel
	peerMsgs                []PeerMsg     // msgs from peers
	currentStep             int           // current step number, start from 0
	stepLength              int           // how many ms to wait after a step
	maxStepLength           int
	minStepLength           int
	onBeforeAddPeer         func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	onAfterAddPeer          func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	onBeforeRemovePeer      func(step int, targetId int, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	onAfterRemovePeer       func(step int, targetId int, peers map[int]Peer, s *LockstepServer)       // lifecycle hook
	onStep                  func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer) // lifecycle hook
	onReceivePeerMsg        func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer)         // lifecycle hook
}

func NewLockStepServer() *LockstepServer {
	return &LockstepServer{
		peers:                   map[int]Peer{},
		stop:                    make(chan bool),
		handleKeyboardInterrupt: false,
		rc:                      make(chan *PeerMsg),
		peerMsgs:                []PeerMsg{},
		currentStep:             0,
		stepLength:              33,  // ~30 step/second
		maxStepLength:           100, // ~10 step/second
		minStepLength:           8,   // ~120 step/second
		onBeforeAddPeer:         func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer) {},
		onAfterAddPeer:          func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer) {},
		onBeforeRemovePeer:      func(step int, targetId int, peers map[int]Peer, s *LockstepServer) {},
		onAfterRemovePeer:       func(step int, targetId int, peers map[int]Peer, s *LockstepServer) {},
		onStep:                  func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer) {},
		onReceivePeerMsg:        func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer) {},
	}
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

// Activate a peer and manage its lifecycle.
func (s *LockstepServer) AddPeer(p Peer) {
	s.onBeforeAddPeer(s.currentStep, p, s.peers, s)

	peerId := 0
	for {
		_, ok := s.peers[peerId]
		if !ok {
			p.Activate(s.rc, peerId)
			s.peers[peerId] = p
			break
		}
		peerId++
	}
	go p.Start()

	s.onAfterAddPeer(s.currentStep, p, s.peers, s)
}

// Close the peer and untrack it.
func (s *LockstepServer) RemovePeer(peerId int) (err error) {
	s.onBeforeRemovePeer(s.currentStep, peerId, s.peers, s)

	if peer, ok := s.peers[peerId]; ok {
		peer.Close()
		delete(s.peers, peerId)
	} else {
		err = errors.New("peer not exist")
	}

	s.onAfterRemovePeer(s.currentStep, peerId, s.peers, s)
	return
}

func (s *LockstepServer) GetPeerCount() int {
	return len(s.peers)
}

func (s *LockstepServer) On(event GameServerLifeCycleEvent, f interface{}) *LockstepServer {
	switch event {
	case BeforeAddPeer:
		s.onBeforeAddPeer = f.(func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer))
	case AfterAddPeer:
		s.onAfterAddPeer = f.(func(step int, newPeer Peer, peers map[int]Peer, s *LockstepServer))
	case BeforeRemovePeer:
		s.onBeforeRemovePeer = f.(func(step int, targetId int, peers map[int]Peer, s *LockstepServer))
	case AfterRemovePeer:
		s.onAfterRemovePeer = f.(func(step int, targetId int, peers map[int]Peer, s *LockstepServer))
	case Step:
		s.onStep = f.(func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer))
	case ReceivePeerMsg:
		s.onReceivePeerMsg = f.(func(step int, peers map[int]Peer, m *PeerMsg, s *LockstepServer))
	}
	return s
}

func (s *LockstepServer) Start() (errs []error) {
	errs = []error{}

	timer := time.NewTimer(time.Duration(s.stepLength))

	// keyboard interrupt handler channel
	kbc := make(chan os.Signal, 1)
	signal.Notify(kbc, os.Interrupt)

	log.Println("lockstep server started, step length:", s.stepLength, "ms")

	loop := true
	for loop {
		select {
		case peerMsg := <-s.rc:
			// accumulate msgs
			s.peerMsgs = append(s.peerMsgs, *peerMsg)
			// handle lifecycle hook
			s.onReceivePeerMsg(s.currentStep, s.peers, peerMsg, s)
		case <-timer.C:
			// handle lifecycle hook
			s.onStep(s.currentStep, s.peers, s.peerMsgs, s)

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
