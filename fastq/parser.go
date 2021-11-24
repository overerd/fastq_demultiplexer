package fastq

import (
	"fastq_demultiplexer/chan_io"
	"fastq_demultiplexer/structs"
	"sync"
	"time"
)

type Read struct {
	Id      string
	Seq     string
	Sup     string
	Quality string
}

type ReadPair struct {
	R1 *Read
	R2 *Read
}

func accumulateReadFromLines(readsChan chan<- Read, input *string, pairBlock *Read, counter *byte) {
	switch *counter {
	case 0:
		(*pairBlock).Id = *input
		*counter++
	case 1:
		(*pairBlock).Seq = *input
		*counter++
	case 2:
		(*pairBlock).Sup = *input
		*counter++
	case 3:
		(*pairBlock).Quality = *input

		readsChan <- *pairBlock
		*counter = 0

		*pairBlock = Read{}
	}
}

func collectBlocks(r1ReadsChan, r2ReadsChan chan<- Read, r1Lines, r2Lines <-chan string) {
	r1BlockLinesRead := byte(0)
	r2BlockLinesRead := byte(0)

	r1Read := Read{}
	r2Read := Read{}

	go func() {
		for {
			select {
			case r1Line := <-r1Lines:
				accumulateReadFromLines(r1ReadsChan, &r1Line, &r1Read, &r1BlockLinesRead)
			}
		}
	}()

	go func() {
		for {
			select {
			case r2Line := <-r2Lines:
				accumulateReadFromLines(r2ReadsChan, &r2Line, &r2Read, &r2BlockLinesRead)
			}
		}
	}()
}

func collectPairReads(r1ReadsChan, r2ReadsChan <-chan Read, readPairsChan chan<- *ReadPair) {
	for {
		r1 := <-r1ReadsChan
		r2 := <-r2ReadsChan

		readPairsChan <- &ReadPair{
			R1: &r1,
			R2: &r2,
		}
	}
}

func ReadR1R2Data(
	options structs.AppOptions,

	r1Input, r2Input *chan_io.IOStringLines,
) (
	readPairsChan chan *ReadPair,
	done chan bool,
) {
	r1ReadsChan := make(chan Read, options.BlockSize)
	r2ReadsChan := make(chan Read, options.BlockSize)

	readPairsChan = make(chan *ReadPair, options.BlockSize)

	done = make(chan bool, 1)

	go collectBlocks(r1ReadsChan, r2ReadsChan, r1Input.Lines, r2Input.Lines)
	go collectPairReads(r1ReadsChan, r2ReadsChan, readPairsChan)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		wg.Wait()

		for {
			time.Sleep(100 * time.Millisecond)

			if len(readPairsChan)+len(r1Input.Lines)+len(r2Input.Lines) == 0 {
				break
			}
		}

		done <- true
	}()

	go func() {
		for {
			select {
			case <-r1Input.Done:
				wg.Done()
			case <-r2Input.Done:
				wg.Done()
			}
		}
	}()

	return
}
