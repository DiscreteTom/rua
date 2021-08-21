package rua

import (
	"errors"
	"log"
	"os"
	"os/signal"
)

type FifoServer struct {
	peers                   map[int]Peer
	stop                    chan bool
	handleKeyboardInterrupt bool
	rc                      chan *PeerMsg // receiver channel
}

func NewFifoServer() *FifoServer {
	return &FifoServer{
		peers:                   map[int]Peer{},
		stop:                    make(chan bool),
		handleKeyboardInterrupt: false,
		rc:                      make(chan *PeerMsg),
	}
}

func (s *FifoServer) SetHandleKeyboardInterrupt(enable bool) *FifoServer {
	s.handleKeyboardInterrupt = enable
	return s
}

// Activate a peer and manage its lifecycle.
func (s *FifoServer) AddPeer(p Peer) {
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
func (s *FifoServer) RemovePeer(peerId int) error {
	if peer, ok := s.peers[peerId]; ok {
		peer.Close()
		delete(s.peers, peerId)
		return nil
	} else {
		return errors.New("peer not exist")
	}
}

func (s *FifoServer) Start(stepHandler func(peers map[int]Peer, m *PeerMsg, s *FifoServer) []error) (errs []error) {
	errs = []error{}

	// keyboard interrupt handler channel
	kbc := make(chan os.Signal, 1)
	signal.Notify(kbc, os.Interrupt)

	log.Println("fifo server started")

	loop := true
	for loop {
		select {
		case peerMsg := <-s.rc:
			// handle step
			errs := stepHandler(s.peers, peerMsg, s)
			if len(errs) != 0 {
				log.Println(errs)
			}
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

func (s *FifoServer) Stop() {
	s.stop <- true
}
