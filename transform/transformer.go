package transform

import (
	"fastq_demultiplexer/barcodes"
	"fastq_demultiplexer/chan_io"
	"fastq_demultiplexer/fastq"
	"fastq_demultiplexer/structs"
	"fastq_demultiplexer/transform/transform_strategies"
	"fmt"
	"time"
)

func TransformData(
	options structs.AppOptions,
	barcodes *barcodes.BarcodeMap,
	barcodeFiles *barcodes.BarcodeFiles,

	r1Input, r2Input *chan_io.IOStringLines,

	done chan<- bool,
) {
	transformStrategy, err := transform_strategies.GetStrategy(options.TransformStrategy)

	if err != nil {
		panic(fmt.Sprintf("Error: %s\n", err.Error()))
	}

	transformer := transformStrategy.BuildTransformer()

	readPairsChan, readPairsDone := fastq.ReadR1R2Data(options, r1Input, r2Input)

	go func() {
		for {
			select {
			case r := <-readPairsDone:
				for len(readPairsChan) > 0 {
					transformer(barcodes, barcodeFiles, <-readPairsChan)
				}

				done <- r

				return
			case block := <-readPairsChan:
				transformer(barcodes, barcodeFiles, block)
			}
		}
	}()

	time.Sleep(time.Second)

	return
}
