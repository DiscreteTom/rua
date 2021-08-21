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
}

// Close the peer and untrack it.
func (s *LockstepServer) RemovePeer(peerId int) error {
	if peer, ok := s.peers[peerId]; ok {
		peer.Close()
		delete(s.peers, peerId)
		return nil
	} else {
		return errors.New("peer not exist")
	}
}

func (s *LockstepServer) Start(stepHandler func(step int, peers map[int]Peer, peerMsgs []PeerMsg, s *LockstepServer) []error) (errs []error) {
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
		case <-timer.C:
			// handle step
			errs := stepHandler(s.currentStep, s.peers, s.peerMsgs, s)
			if len(errs) != 0 {
				log.Println(errs)
			}
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
