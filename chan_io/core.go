package chan_io

import (
	"sync"
	"time"
)

type IOStringLines struct {
	Lines  chan string
	Done   chan bool
	IsDone bool
}

func NewIOStringLines(size uint) *IOStringLines {
	return &IOStringLines{
		Lines: make(chan string, size),
		Done:  make(chan bool, 1),
	}
}

func (io *IOStringLines) EnsureWriteWithClose(wg *sync.WaitGroup) {
	for len(io.Lines) > 0 {
		time.Sleep(time.Second)
	}

	if wg != nil {
		wg.Done()
	}

	io.Done <- true
}

func (io *IOStringLines) closeLines() {
	close(io.Lines)
}

func (io *IOStringLines) closeDone() {
	close(io.Done)
}

func (io *IOStringLines) Close() {
	io.closeLines()
	io.closeDone()
}
