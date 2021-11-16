package main

import (
	"github.com/DiscreteTom/rua"
)

func main() {
	stdio_node := rua.DefaultStdioNode()
	stdio := stdio_node.Handle()
	stdio_node.OnInput(func(b []byte) { stdio.Write(b) }).Go()

	rua.NewCtrlc().OnSignal(func() {
		stdio.Stop()
	}).Wait()
}
