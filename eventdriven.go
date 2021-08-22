package rua

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"time"
)

type EventDrivenServer struct {
	stop                     chan bool
	handleKeyboardInterrupt  bool
	peers                    map[int]Peer                                                 // peer id starts from 0
	beforeAddPeerHandler     func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	afterAddPeerHandler      func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	beforeRemovePeerHandler  func(targetId int, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	afterRemovePeerHandler   func(targetId int, peers map[int]Peer, s *EventDrivenServer) // lifecycle hook
	beforeProcPeerMsgHandler func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer)   // lifecycle hook
	onPeerMsgHandler         func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer)   // lifecycle hook
}

func NewEventDrivenServer() *EventDrivenServer {
	return &EventDrivenServer{
		stop:                     make(chan bool),
		handleKeyboardInterrupt:  false,
		peers:                    map[int]Peer{},
		beforeAddPeerHandler:     func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) {},
		afterAddPeerHandler:      func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer) {},
		beforeRemovePeerHandler:  func(targetId int, peers map[int]Peer, s *EventDrivenServer) {},
		afterRemovePeerHandler:   func(targetId int, peers map[int]Peer, s *EventDrivenServer) {},
		beforeProcPeerMsgHandler: func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer) {},
		onPeerMsgHandler:         func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer) {},
	}
}

func (s *EventDrivenServer) SetHandleKeyboardInterrupt(enable bool) *EventDrivenServer {
	s.handleKeyboardInterrupt = enable
	return s
}

// Activate a peer, allocate a peerId and manage the peer's lifecycle.
func (s *EventDrivenServer) AddPeer(p Peer) {
	// allocate a peerId
	peerId := 0
	for {
		_, ok := s.peers[peerId]
		if !ok {
			break
		}
		peerId++
	}

	p.Activate(peerId)
	s.beforeAddPeerHandler(p, s.peers, s)

	// add the peer and start the peer
	s.peers[peerId] = p
	go p.Start()

	s.afterAddPeerHandler(p, s.peers, s)

}

// Close the peer and untrack it.
func (s *EventDrivenServer) RemovePeer(peerId int) (err error) {
	s.beforeRemovePeerHandler(peerId, s.peers, s)

	if peer, ok := s.peers[peerId]; ok {
		peer.Close()
		delete(s.peers, peerId)
	} else {
		err = errors.New("peer not exist")
	}

	s.afterRemovePeerHandler(peerId, s.peers, s)
	return
}

func (s *EventDrivenServer) GetPeerCount() int {
	return len(s.peers)
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, but `peers` not contains the new peer.
func (s *EventDrivenServer) BeforeAddPeer(f func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer)) *EventDrivenServer {
	s.beforeAddPeerHandler = f
	return s
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, and `peers` contains the new peer.
func (s *EventDrivenServer) AfterAddPeer(f func(newPeer Peer, peers map[int]Peer, s *EventDrivenServer)) *EventDrivenServer {
	s.afterAddPeerHandler = f
	return s
}

// Register lifecycle hook.
func (s *EventDrivenServer) BeforeRemovePeer(f func(targetId int, peers map[int]Peer, s *EventDrivenServer)) *EventDrivenServer {
	s.beforeRemovePeerHandler = f
	return s
}

// Register lifecycle hook.
func (s *EventDrivenServer) AfterRemovePeer(f func(targetId int, peers map[int]Peer, s *EventDrivenServer)) *EventDrivenServer {
	s.afterRemovePeerHandler = f
	return s
}

// Register lifecycle hook.
// You can modify or enrich the peer message before process it.
func (s *EventDrivenServer) BeforeProcPeerMsg(f func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer)) *EventDrivenServer {
	s.beforeProcPeerMsgHandler = f
	return s
}

// Register lifecycle hook.
func (s *EventDrivenServer) OnPeerMsg(f func(peers map[int]Peer, m *PeerMsg, s *EventDrivenServer)) *EventDrivenServer {
	s.onPeerMsgHandler = f
	return s
}

func (s *EventDrivenServer) Start() (errs []error) {
	errs = []error{}

	// keyboard interrupt handler channel
	kbc := make(chan os.Signal, 1)
	signal.Notify(kbc, os.Interrupt)

	log.Println("eventdriven server started")

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
	// this hook can modify peerMsg before append
	s.beforeProcPeerMsgHandler(s.peers, &peerMsg, s)

	// handle lifecycle hook
	s.onPeerMsgHandler(s.peers, &peerMsg, s)
}
