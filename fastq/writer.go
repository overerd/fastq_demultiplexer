package fastq

import (
	"bufio"
	"compress/gzip"
	"fastq_demultiplexer/chan_io"
	"fastq_demultiplexer/structs"
	"fmt"
	"io"
	"os"
	"sync"
)

func WriteFile(appOptions *structs.AppOptions, fileOptions structs.FileOptions, input *chan_io.IOStringLines, wg *sync.WaitGroup) (err error) {
	file, err := os.OpenFile(fileOptions.Path, os.O_WRONLY|os.O_CREATE, 0755)

	if err != nil {
		return
	}

	if fileOptions.Debug {
		fmt.Printf("[>] %s\n", fileOptions.Filename)
	}

	var innerWriter io.WriteCloser

	if fileOptions.UseCompression {
		innerWriter, err = gzip.NewWriterLevel(file, fileOptions.CompressionLevel)

		if err != nil {
			return
		}
	} else {
		innerWriter = file
	}

	buffWriter := bufio.NewWriterSize(innerWriter, int(appOptions.BufferSize))

	bytes := uint64(0)

	deferredClose := func() {
		err = buffWriter.Flush()

		if err != nil {
			fmt.Printf("\nError: %s\n", err.Error())
		}

		if file != innerWriter {
			err = innerWriter.Close()

			if err != nil {
				fmt.Printf("\nError: %s\n", err.Error())
			}
		}

		err = file.Close()

		if err != nil {
			fmt.Printf("\nError: %s\n", err.Error())
		}

		wg.Done()
	}

	writeLine := func(line *string) {
		b, lineErr := buffWriter.Write([]byte(*line))

		if lineErr != nil {
			panic(lineErr)
		} else {
			bytes += uint64(b)
		}
	}

	go func() {
		go func() {
			for {
				select {
				case line := <-input.Lines:
					writeLine(&line)
				}
			}
		}()

		go func() {
			defer deferredClose()

			for {
				select {
				case <-input.Done:
					for len(input.Lines) > 0 {
						line := <-input.Lines

						writeLine(&line)
					}

					return
				}
			}
		}()
	}()

	return
}
