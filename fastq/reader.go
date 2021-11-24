package fastq

import (
	"bufio"
	"compress/gzip"
	"fastq_demultiplexer/chan_io"
	"fastq_demultiplexer/structs"
	"fmt"
	"io"
	"os"
	"time"
)

func ReadFile(options structs.AppOptions, fileOptions structs.FileOptions, output *chan_io.IOStringLines) (err error) {
	file, err := os.OpenFile(fileOptions.Path, os.O_RDONLY, 0755)

	if err != nil {
		return
	}

	throwError := func(err error) {
		panic(fmt.Sprintf("read error '%s': %s", fileOptions.Filename, err.Error()))
	}

	var innerReader io.ReadCloser

	if fileOptions.UseCompression {
		innerReader, err = gzip.NewReader(file)

		if err != nil {
			return
		}
	} else {
		innerReader = file
	}

	stat, err := file.Stat()

	if err != nil {
		return
	}

	bytes := 0
	fileSize := stat.Size()

	fmt.Printf("[<] %s (%0.2f mbytes)\n", fileOptions.Filename, float64(fileSize)/1024/1024)

	var compressionString string

	if fileOptions.UseCompression {
		compressionString = "compressed "
	} else {
		compressionString = ""
	}

	printProgress := func() {
		fmt.Printf(
			" [!] %s | ~~%0.1f | read %0.2f mbytes (out of %s%0.2f mbytes)\n",
			fileOptions.Filename,
			float64(bytes)/float64(fileSize)*100,
			float64(bytes)/1024/1024,
			compressionString,
			float64(fileSize)/1024/1024,
		)
	}

	go func() {
		for {
			time.Sleep(time.Second * 60)

			printProgress()
		}
	}()

	go func() {
		defer func(file *os.File) {
			err := file.Close()

			if err != nil {
				throwError(err)
			} else {
				output.Done <- true
			}
		}(file)

		if fileOptions.UseCompression {
			defer func(gzFile io.ReadCloser) {
				err := gzFile.Close()

				if err != nil {
					throwError(err)
				}
			}(innerReader)
		}

		scanner := bufio.NewScanner(innerReader)

		scanner.Split(bufio.ScanWords)

		buffer := make([]byte, options.BufferSize)

		scanner.Buffer(buffer, int(options.BufferSize)*10)

		for scanner.Scan() {
			lineBytes := scanner.Bytes()

			bytes += len(lineBytes)

			output.Lines <- string(lineBytes)
		}

		printProgress()

		if err := scanner.Err(); err != nil {
			throwError(err)
		}
	}()

	return
}
