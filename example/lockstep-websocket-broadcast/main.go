package main

import (
	"DiscreteTom/rua"
	"DiscreteTom/rua/plugin/network/websocket"
	"fmt"
	"log"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockStepServer().SetHandleKeyboardInterrupt(true)

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

func broadcastStepHandler(step int, peers map[int]rua.Peer, msgs []rua.PeerMsg, _ *rua.LockstepServer) (errs []error) {
	// compact commands in one byte array
	result := []byte(fmt.Sprintf("step: %d\n", step))
	for _, msg := range msgs {
		result = append(result, []byte(fmt.Sprintf("from %d:\n", msg.PeerId))...)
		result = append(result, msg.Data...)
		result = append(result, '\n')
	}
	// broadcast to everyone
	for _, p := range peers {
		go p.Write(result)
	}
	return
}
