package main

import "github.com/DiscreteTom/rua"

func main() {
	file, _ := rua.DefaultFileNode().Filename("log.txt").Go()

	stdio_node := rua.DefaultStdioNode()
	stdio := stdio_node.Handle()

	stdio_node.OnInput(func(b []byte) {
		file.WriteThen(b, func(e error) {
			if e == nil {
				stdio.Write([]byte("ok"))
			} else {
				stdio.Write([]byte("err"))
			}
		})
	}).Go()

	rua.NewCtrlc().OnSignal(func() {
		file.StopThen(func(e error) {
			stdio.Stop()
		})
	}).Wait()
}
