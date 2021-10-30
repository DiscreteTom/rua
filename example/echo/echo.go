package main

import (
	"github.com/DiscreteTom/rua"
)

func main() {
	node := rua.DefaultStdioNode()
	handle := node.Handle()
	node.OnMsg(func(b []byte) { handle.Write(b) }).Go()

	rua.NewCtrlc().OnSignal(func() {
		handle.Stop()
	}).Wait()
}
