package main

import (
	"strconv"

	"github.com/DiscreteTom/rua"
)

func main() {
	state := []byte{}

	stdio := rua.DefaultStdioNode().OnMsg(func(b []byte) {
		state = append(state, b...)
	}).Go()

	ls := rua.NewLockstep().StepLengthMs(1000).OnStep(func(u uint64) {
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
