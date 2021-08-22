package rua

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"time"
)

type EventDrivenServer struct {
	stop                    chan bool
	handleKeyboardInterrupt bool
	peers                   map[int]Peer                                                 // peer id starts from 0
	onBeforeAddPeer         func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	onAfterAddPeer          func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	onBeforeRemovePeer      func(targetId int, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	onAfterRemovePeer       func(targetId int, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	onReceivePeerMsg        func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer)   // lifecycle hook
}

func NewEventDrivenServer() *EventDrivenServer {
	return &EventDrivenServer{
		stop:                    make(chan bool),
		handleKeyboardInterrupt: false,
		peers:                   map[int]Peer{},
		onBeforeAddPeer:         func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) {},
		onAfterAddPeer:          func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) {},
		onBeforeRemovePeer:      func(targetId int, peers map[int]Peer, s *EventDrivenServer) {},
		onAfterRemovePeer:       func(targetId int, peers map[int]Peer, s *EventDrivenServer) {},
		onReceivePeerMsg:        func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer) {},
	}
}

func (s *EventDrivenServer) SetHandleKeyboardInterrupt(enable bool) *EventDrivenServer {
	s.handleKeyboardInterrupt = enable
	return s
}

// Activate a peer and manage its lifecycle.
func (s *EventDrivenServer) AddPeer(p Peer) {
	s.onBeforeAddPeer(p, s.peers, s)

	peerId := 0
	for {
		_, ok := s.peers[peerId]
		if !ok {
			p.Activate(peerId)
			s.peers[peerId] = p
			break
		}
		peerId++
	}
	go p.Start()

	s.onAfterAddPeer(p, s.peers, s)
}

// Close the peer and untrack it.
func (s *EventDrivenServer) RemovePeer(peerId int) (err error) {
	s.onBeforeRemovePeer(peerId, s.peers, s)

	if peer, ok := s.peers[peerId]; ok {
		peer.Close()
		delete(s.peers, peerId)
	} else {
		err = errors.New("peer not exist")
	}

	s.onAfterRemovePeer(peerId, s.peers, s)
	return
}

func (s *EventDrivenServer) GetPeerCount() int {
	return len(s.peers)
}

func (s *EventDrivenServer) On(event GameServerLifeCycleEvent, f interface{}) *EventDrivenServer {
	switch event {
	case BeforeAddPeer:
		s.onBeforeAddPeer = f.(func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer))
	case AfterAddPeer:
		s.onAfterAddPeer = f.(func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer))
	case BeforeRemovePeer:
		s.onBeforeRemovePeer = f.(func(targetId int, peers map[int]Peer, s *EventDrivenServer))
	case AfterRemovePeer:
		s.onAfterRemovePeer = f.(func(targetId int, peers map[int]Peer, s *EventDrivenServer))
	case ReceivePeerMsg:
		s.onReceivePeerMsg = f.(func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer))
	}
	return s
}

func (s *EventDrivenServer) Start() (errs []error) {
	errs = []error{}

	// keyboard interrupt handler channel
	kbc := make(chan os.Signal, 1)
	signal.Notify(kbc, os.Interrupt)

	log.Println("fifo server started")

	loop := true
	for loop {
		select {
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
	return
}

func (s *EventDrivenServer) Stop() {
	s.stop <- true
}

func (s *EventDrivenServer) AppendPeerMsg(peerId int, d []byte) {
	peerMsg := PeerMsg{PeerId: peerId, Data: d, Time: time.Now()}
	// handle lifecycle hook
	s.onReceivePeerMsg(s.peers, &peerMsg, s)
}
