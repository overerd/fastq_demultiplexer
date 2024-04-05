package transform_strategies

import (
	"fastq_demultiplexer/barcodes"
	"fastq_demultiplexer/fastq"
	"fmt"
)

type SimpleBaseTransformStrategy struct{}

func (s *SimpleBaseTransformStrategy) GetSchema() []string { return []string{"R1", "R2"} }
func (s *SimpleBaseTransformStrategy) BuildTransformer() func(barcodes *barcodes.BarcodeMap, barcodeFiles *barcodes.BarcodeFiles, block *fastq.ReadPair) {
	return func(
		barcodes *barcodes.BarcodeMap,
		barcodeFiles *barcodes.BarcodeFiles,
		block *fastq.ReadPair,
	) {
		r1Id := block.R1.Id[:len(block.R1.Id)-2] // trim last two characters of orientation part
		r2Id := block.R2.Id[:len(block.R2.Id)-2]

		if r1Id != r2Id {
			panic(
				fmt.Sprintf(
					"R1-R2 Ids inconsistency ('%s' vs '%s'). Try sorting.",
					r1Id,
					r2Id,
				),
			)

			return
		}

		barcodeSeq := block.R1.Seq[len(block.R1.Seq)-8:]
		reversedBarcodeSeq := ReverseSequence(barcodeSeq)

		if barcodeId, ok := (*barcodes)[reversedBarcodeSeq]; ok {
			if chanIO, ok := (*barcodeFiles)[barcodeId]; ok {
				r1Id = fmt.Sprintf("%s 1:N:0:%s", block.R1.Id, barcodeSeq)
				r2Id = fmt.Sprintf("%s 2:N:0:%s", block.R2.Id, barcodeSeq)

				r1Seq := block.R1.Seq[:28]
				r1Quality := block.R1.Quality[:28]

				r2Len := len(block.R2.Seq) - 9

				if r2Len <= 0 {
					r2Len = len(block.R2.Seq)
				}

				r2Seq := block.R2.Seq[:r2Len]
				r2Quality := block.R2.Quality[:r2Len]

				chanIO.Files["R1"].ChanIO.Lines <- fmt.Sprintf(
					"%s\n%s\n%s\n%s\n",
					r1Id, r1Seq, block.R1.Sup, r1Quality,
				)

				chanIO.Files["R2"].ChanIO.Lines <- fmt.Sprintf(
					"%s\n%s\n%s\n%s\n",
					r2Id, r2Seq, block.R2.Sup, r2Quality,
				)
			}
		}
	}
}
