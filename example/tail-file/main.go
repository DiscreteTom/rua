package main

import (
	"strconv"

	"github.com/DiscreteTom/rua"
)

func main() {
	const FILENAME = "log.txt"

	file, _ := rua.DefaultFileNode().Filename(FILENAME).Go()

	// ticker will write current tick count to file
	ticker, _ := rua.DefaultTicker().OnTick(func(u uint64) {
		file.Write([]byte(strconv.FormatUint(u, 10)))
	}).Go()

	// tail file to stdout
	stdio := rua.DefaultStdioNode().Go()
	tail, _ := rua.NewTailNode(FILENAME).OnNewLine(func(b []byte) {
		stdio.Write(b)
	}).Go()

	rua.NewCtrlc().OnSignal(func() {
		ticker.Stop()
		tail.Stop()
		file.Stop()
		stdio.Stop()
	}).Wait()
}
