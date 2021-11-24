package transform_strategies

import (
	"fastq_demultiplexer/barcodes"
	"fastq_demultiplexer/fastq"
	"fmt"
)

type BaseTransformStrategy struct{}

func (s *BaseTransformStrategy) GetSchema() []string { return []string{"R1", "R2", "I1"} }
func (s *BaseTransformStrategy) BuildTransformer() func(barcodes *barcodes.BarcodeMap, barcodeFiles *barcodes.BarcodeFiles, shard *fastq.ReadPair) {
	return func(
		barcodes *barcodes.BarcodeMap,
		barcodeFiles *barcodes.BarcodeFiles,
		shard *fastq.ReadPair,
	) {
		r1Id := shard.R1.Id[:len(shard.R1.Id)-2] // trim last two characters of orientation part
		r2Id := shard.R2.Id[:len(shard.R2.Id)-2]

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

		barcodeSeq := shard.R1.Seq[len(shard.R1.Seq)-8:]
		reversedBarcodeSeq := ReverseSequence(barcodeSeq)

		if barcodeId, ok := (*barcodes)[reversedBarcodeSeq]; ok {
			if chanIO, ok := (*barcodeFiles)[barcodeId]; ok {
				i1Id := fmt.Sprintf("%s 2:N:0:%s", r2Id, barcodeSeq)
				r1Id = fmt.Sprintf("%s 1:N:0:%s", shard.R1.Id, barcodeSeq)
				r2Id = fmt.Sprintf("%s 2:N:0:%s", shard.R2.Id, barcodeSeq)

				r1Seq := shard.R1.Seq[:28]
				r1Quality := shard.R1.Quality[:28]

				r2Len := len(shard.R2.Seq) - 9

				if r2Len <= 0 {
					r2Len = len(shard.R2.Seq)
				}

				r2Seq := shard.R2.Seq[:r2Len]
				r2Quality := shard.R2.Quality[:r2Len]

				i1Quality := r1Quality[len(r1Quality)-8:]

				chanIO.Files["R1"].ChanIO.Lines <- fmt.Sprintf(
					"%s\n%s\n%s\n%s\n",
					r1Id, r1Seq, shard.R1.Sup, r1Quality,
				)

				chanIO.Files["R2"].ChanIO.Lines <- fmt.Sprintf(
					"%s\n%s\n%s\n%s\n",
					r2Id, r2Seq, shard.R2.Sup, r2Quality,
				)

				chanIO.Files["I1"].ChanIO.Lines <- fmt.Sprintf(
					"%s\n%s\n+\n%s\n",
					i1Id, barcodeSeq, i1Quality,
				)
			}
		}
	}
}
