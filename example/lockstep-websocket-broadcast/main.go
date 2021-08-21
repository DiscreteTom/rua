package main

import (
	"DiscreteTom/rua/pkg/lockstep"
	"DiscreteTom/rua/pkg/model"
	"DiscreteTom/rua/plugin/network/websocket"
	"fmt"
	"log"
)

func main() {
	errChan := make(chan error)
	s := lockstep.NewLockStepServer().SetHandleKeyboardInterrupt(true)

	ws := websocket.NewWebsocketListener(":8080", s)
	go func() {
		errChan <- ws.Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start(broadcastStepHandler)
	}()

	select {
	case err := <-errChan:
		log.Println(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			log.Println(errs)
		}
		break
	}
}

func broadcastStepHandler(step int, peers map[int]model.Peer, commands map[int][]model.PeerCommand, _ *lockstep.LockstepServer) (errs []error) {
	// compact commands in one byte array
	msg := []byte(fmt.Sprintf("step: %d\n", step))
	for id, cmds := range commands {
		msg = append(msg, []byte(fmt.Sprintf("from %d:\n", id))...)
		for _, c := range cmds {
			msg = append(msg, c.Data...)
		}
		msg = append(msg, '\n')
	}
	// broadcast to everyone
	for _, p := range peers {
		go p.Write(msg)
	}
	return
}
