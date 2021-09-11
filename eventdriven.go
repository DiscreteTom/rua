package rua

import (
	"errors"
	"sync"
	"time"
)

type EventDrivenServer struct {
	name                     string
	stop                     chan bool
	peers                    map[int]Peer       // peer id starts from 0
	peerLock                 *sync.Mutex        // prevent concurrent access
	beforeAddPeerHandler     func(newPeer Peer) // lifecycle hook
	afterAddPeerHandler      func(newPeer Peer) // lifecycle hook
	beforeRemovePeerHandler  func(targetId int) // lifecycle hook
	afterRemovePeerHandler   func(targetId int) // lifecycle hook
	beforeProcPeerMsgHandler func(m *PeerMsg)   // lifecycle hook
	onPeerMsgHandler         func(m *PeerMsg)   // lifecycle hook
	logger                   Logger
}

func NewEventDrivenServer() (*EventDrivenServer, error) {
	return &EventDrivenServer{
		name:                     "EventDrivenServer",
		stop:                     make(chan bool),
		peers:                    map[int]Peer{},
		peerLock:                 &sync.Mutex{},
		beforeAddPeerHandler:     func(newPeer Peer) {},
		afterAddPeerHandler:      func(newPeer Peer) {},
		beforeRemovePeerHandler:  func(targetId int) {},
		afterRemovePeerHandler:   func(targetId int) {},
		beforeProcPeerMsgHandler: func(m *PeerMsg) {},
		onPeerMsgHandler:         func(m *PeerMsg) {},
		logger:                   DefaultLogger(),
	}, nil
}

func (s *EventDrivenServer) SetName(n string) {
	s.name = n
}

func (s *EventDrivenServer) SetLogger(l Logger) {
	s.logger = l
}

func (s *EventDrivenServer) Logger() Logger {
	return s.logger
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
	s.beforeAddPeerHandler(p)
	s.peers[peerId] = p

	s.peerLock.Unlock()
	go p.Start()

	s.afterAddPeerHandler(p)
	return peerId
}

// Close the peer and untrack it. Return err if peer not exist.
func (s *EventDrivenServer) RemovePeer(peerId int) (err error) {
	s.beforeRemovePeerHandler(peerId)

	s.peerLock.Lock()
	if peer, ok := s.peers[peerId]; ok {
		if err := peer.Close(); err != nil {
			s.logger.Errorf("rua.%s.RemovePeer: %s", s.name, err)
		}
		delete(s.peers, peerId)
	} else {
		err = errors.New("peer not exist")
	}
	s.peerLock.Unlock()

	s.afterRemovePeerHandler(peerId)
	return
}

// Thread safe. Do NOT AddPeer or RemovePeer in f.
func (s *EventDrivenServer) ForEachPeer(f func(id int, peer Peer)) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	for i, p := range s.peers {
		f(i, p)
	}
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, but the new peer is not started, and `peers` does not contain the new peer.
// This hook won't be triggered concurrently.
func (s *EventDrivenServer) BeforeAddPeer(f func(newPeer Peer)) {
	s.beforeAddPeerHandler = f
}

// Register lifecycle hook.
// At this time the new peer's id has been allocated, the peer is started and `peers` contains the new peer.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) AfterAddPeer(f func(newPeer Peer)) {
	s.afterAddPeerHandler = f
}

// Register lifecycle hook.
// The target peer may has been closed.
// The target peer may not exist.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) BeforeRemovePeer(f func(targetId int)) {
	s.beforeRemovePeerHandler = f
}

// Register lifecycle hook.
// The target peer may not exist.
// If it exists, it must been closed, and been removed from `peers`.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) AfterRemovePeer(f func(targetId int)) {
	s.afterRemovePeerHandler = f
}

// Register lifecycle hook.
// You can modify or enrich the peer message before process it.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) BeforeProcPeerMsg(f func(m *PeerMsg)) {
	s.beforeProcPeerMsgHandler = f
}

// Register lifecycle hook.
// This hook may be triggered concurrently.
func (s *EventDrivenServer) OnPeerMsg(f func(m *PeerMsg)) {
	s.onPeerMsgHandler = f
}

// Return errors from peer.Close() when stop the server.
func (s *EventDrivenServer) Start() (errs []error) {
	errs = []error{}

	s.logger.Infof("%s started", s.name)

	<-s.stop

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

func (s *EventDrivenServer) AppendPeerMsg(p Peer, d []byte) {
	peerMsg := PeerMsg{Peer: p, Data: d, Time: time.Now()}

	// handle lifecycle hook
	// this hook can modify peerMsg before append
	s.beforeProcPeerMsgHandler(&peerMsg)

	// handle lifecycle hook
	s.onPeerMsgHandler(&peerMsg)
}
