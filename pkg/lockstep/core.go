package lockstep

import (
	"DiscreteTom/rua/pkg/model"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

type lockstepServer struct {
	peers         map[int]model.Peer
	rc            chan *model.PeerMsg // receiver channel
	commands      map[int][]byte      // commands from peers
	stepLength    int                 // how many ms to wait after a step
	currentStep   int
	maxStepLength int
}

func NewLockStepServer() *lockstepServer {
	return &lockstepServer{
		peers:         map[int]model.Peer{},
		rc:            make(chan *model.PeerMsg),
		commands:      map[int][]byte{},
		stepLength:    30,
		currentStep:   0,
		maxStepLength: 1000,
	}
}

func (s *lockstepServer) SetStepLength(stepLength int) (err error) {
	if stepLength <= s.maxStepLength {
		s.stepLength = stepLength
	} else {
		err = fmt.Errorf("step length %d exceed max step length %d", stepLength, s.maxStepLength)
	}
	return
}

func (s *lockstepServer) SetMaxStepLength(maxStepLength int) *lockstepServer {
	s.maxStepLength = maxStepLength
	return s
}

func (s *lockstepServer) AddPeer(p model.Peer) {
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

func (s *lockstepServer) RemovePeer(peerId int) error {
	if peer, ok := s.peers[peerId]; ok {
		peer.Close()
		delete(s.peers, peerId)
		return nil
	} else {
		return errors.New("peer not exist")
	}
}

func (s *lockstepServer) Start(stepHandler func(step int, peers map[int]model.Peer, commands map[int][]byte) []error) (errs []error) {
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
			// accumulate commands
			s.commands[peerMsg.PeerId] = append(s.commands[peerMsg.PeerId], peerMsg.Data...)
		case <-timer.C:
			// handle step
			stepHandler(s.currentStep, s.peers, s.commands)
			s.currentStep++
			// reset commands
			s.commands = map[int][]byte{}
			// reset timer
			timer = time.NewTimer(time.Duration(s.stepLength) * time.Millisecond)
		case <-kbc:
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
