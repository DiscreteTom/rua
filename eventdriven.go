package rua

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

type EventDrivenServer struct {
	stop                     chan bool
	handleKeyboardInterrupt  bool
	peers                    map[int]Peer                             // peer id starts from 0
	peerLock                 sync.Mutex                               // prevent concurrent access
	beforeAddPeerHandler     func(newPeer Peer, s *EventDrivenServer) // lifecycle hook
	afterAddPeerHandler      func(newPeer Peer, s *EventDrivenServer) // lifecycle hook
	beforeRemovePeerHandler  func(targetId int, s *EventDrivenServer) // lifecycle hook
	afterRemovePeerHandler   func(targetId int, s *EventDrivenServer) // lifecycle hook
	beforeProcPeerMsgHandler func(m *PeerMsg, s *EventDrivenServer)   // lifecycle hook
	onPeerMsgHandler         func(m *PeerMsg, s *EventDrivenServer)   // lifecycle hook
	logger                   Logger
}

func NewEventDrivenServer() *EventDrivenServer {
	return &EventDrivenServer{
		stop:                     make(chan bool),
		handleKeyboardInterrupt:  false,
		peers:                    map[int]Peer{},
		peerLock:                 sync.Mutex{},
		beforeAddPeerHandler:     func(newPeer Peer, s *EventDrivenServer) {},
		afterAddPeerHandler:      func(newPeer Peer, s *EventDrivenServer) {},
		beforeRemovePeerHandler:  func(targetId int, s *EventDrivenServer) {},
		afterRemovePeerHandler:   func(targetId int, s *EventDrivenServer) {},
		beforeProcPeerMsgHandler: func(m *PeerMsg, s *EventDrivenServer) {},
		onPeerMsgHandler:         func(m *PeerMsg, s *EventDrivenServer) {},
		logger:                   GetDefaultLogger(),
	}
}

func (s *EventDrivenServer) WithLogger(l Logger) *EventDrivenServer {
	s.logger = l
	return s
}

func (s *EventDrivenServer) GetLogger() Logger {
	return s.logger
}

func (s *EventDrivenServer) SetHandleKeyboardInterrupt(enable bool) *EventDrivenServer {
	s.handleKeyboardInterrupt = enable
	return s
}

// Activate a peer, allocate a peerId and manage the peer's lifecycle.
func (s *EventDrivenServer) AddPeer(p Peer) int {
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
	s.beforeAddPeerHandler(p, s)
	s.peers[peerId] = p

	s.peerLock.Unlock()
	go p.Start()

	s.afterAddPeerHandler(p, s)
	return peerId
}

// Close the peer and untrack it. Return err if peer not exist.
func (s *EventDrivenServer) RemovePeer(peerId int) (err error) {
	s.beforeRemovePeerHandler(peerId, s)

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

	s.afterRemovePeerHandler(peerId, s)
	return
}

func (s *EventDrivenServer) GetPeerCount() int {
	return len(s.peers)
}

func (s *EventDrivenServer) GetPeers() map[int]Peer {
	return s.peers
}

// Return a peer or nil
func (s *EventDrivenServer) GetPeer(id int) Peer {
	if p, ok := s.peers[id]; ok {
		return p
	}
	return nil
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, but `peers` not contains the new peer.
// This hook won't be triggered concurrently.
func (s *EventDrivenServer) BeforeAddPeer(f func(newPeer Peer, s *EventDrivenServer)) *EventDrivenServer {
	s.beforeAddPeerHandler = f
	return s
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, and `peers` contains the new peer.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) AfterAddPeer(f func(newPeer Peer, s *EventDrivenServer)) *EventDrivenServer {
	s.afterAddPeerHandler = f
	return s
}

// Register lifecycle hook.
// The target peer may has been closed.
// The target peer may not exist.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) BeforeRemovePeer(f func(targetId int, s *EventDrivenServer)) *EventDrivenServer {
	s.beforeRemovePeerHandler = f
	return s
}

// Register lifecycle hook.
// The target peer may not exist.
// If it exists, it must been closed, and been removed from `peers`.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) AfterRemovePeer(f func(targetId int, s *EventDrivenServer)) *EventDrivenServer {
	s.afterRemovePeerHandler = f
	return s
}

// Register lifecycle hook.
// You can modify or enrich the peer message before process it.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) BeforeProcPeerMsg(f func(m *PeerMsg, s *EventDrivenServer)) *EventDrivenServer {
	s.beforeProcPeerMsgHandler = f
	return s
}

// Register lifecycle hook.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) OnPeerMsg(f func(m *PeerMsg, s *EventDrivenServer)) *EventDrivenServer {
	s.onPeerMsgHandler = f
	return s
}

// Return errors from peer.Close() when stop the server.
func (s *EventDrivenServer) Start() (errs []error) {
	errs = []error{}

	// keyboard interrupt handler channel
	kbc := make(chan os.Signal, 1)
	signal.Notify(kbc, os.Interrupt)

	s.logger.Info("eventdriven server started")

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
	s.beforeProcPeerMsgHandler(&peerMsg, s)

	// handle lifecycle hook
	s.onPeerMsgHandler(&peerMsg, s)
}
