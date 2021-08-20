package main

import (
	"DiscreteTom/rua/pkg/model"
	"fmt"
)

func broadcastStepHandler(step int, peers map[int]model.Peer, commands map[int][]byte) (errs []error) {
	// compact commands in one byte array
	msg := []byte(fmt.Sprintf("step: %d\n", step))
	for id, m := range commands {
		msg = append(msg, []byte(fmt.Sprintf("from %d:\n", id))...)
		msg = append(msg, m...)
		msg = append(msg, '\n')
	}
	// broadcast to everyone
	for _, p := range peers {
		go p.Write(msg)
	}
	return
}
