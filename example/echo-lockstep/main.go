package main

import (
	"strconv"
	"sync"

	"github.com/DiscreteTom/rua"
)

func main() {
	lock := &sync.Mutex{}
	state := []byte{}

	stdio := rua.DefaultStdioNode().OnInput(func(b []byte) {
		lock.Lock()
		defer lock.Unlock()
		state = append(state, b...)
	}).Go()

	ls, _ := rua.DefaultTicker().OnTick(func(u uint64) {
		lock.Lock()
		defer lock.Unlock()
		// construct output
		buffer := []byte{}
		buffer = append(buffer, []byte(strconv.FormatUint(u, 10))...)
		buffer = append(buffer, []byte("\n")...)
		buffer = append(buffer, state...)
		// write
		stdio.Write(buffer)
		// reset state
		state = []byte{}
	}).Go()

	rua.NewCtrlc().OnSignal(func() {
		ls.Stop()
		stdio.Stop()
	}).Wait()
}
