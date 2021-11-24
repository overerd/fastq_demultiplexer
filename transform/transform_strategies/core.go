package transform_strategies

import (
	"errors"
	"fastq_demultiplexer/barcodes"
	"fastq_demultiplexer/fastq"
	"fastq_demultiplexer/misc"
	"fmt"
	"strings"
)

var transformStrategies map[string]interface{}
var supportedStrategiesList []string

type TransformStrategy interface {
	GetSchema() []string
	BuildTransformer() func(barcodes *barcodes.BarcodeMap, barcodeFiles *barcodes.BarcodeFiles, shard *fastq.ReadPair)
}

func ReverseSequence(seq string) string {
	res := make([]string, len(seq))

	l := len(seq) - 1

	for i, c := range seq {
		ni := l - i

		switch c {
		case 'A':
			res[ni] = "T"
		case 'T':
			res[ni] = "A"
		case 'G':
			res[ni] = "C"
		case 'C':
			res[ni] = "G"
		default:
			res[ni] = string(c)
		}
	}

	return strings.Join(res, "")
}

func InitStrategies() {
	transformStrategies = map[string]interface{}{
		"10x":          &BaseTransformStrategy{},
		"10x_no_index": &SimpleBaseTransformStrategy{},
	}

	supportedStrategiesList = misc.ExtractMapStringKeys(&transformStrategies)

	return
}

func GetAllStrategies() (strategiesMap map[string]interface{}, list []string, err error) {
	if transformStrategies == nil {
		err = errors.New("unable to access uninitialized map of strategies, run InitStrategies() first")
	}

	return transformStrategies, supportedStrategiesList, err
}

func GetStrategy(strategyName string) (strategy TransformStrategy, err error) {
	transformStrategies, supportedStrategiesList, err := GetAllStrategies()

	if err != nil {
		return
	}

	if res, ok := transformStrategies[strategyName]; ok {
		strategy = res.(TransformStrategy)
	} else {
		err = errors.New(fmt.Sprintf("unknown strategy '%s' (supported strategies: %s)", strategyName, strings.Join(supportedStrategiesList, ", ")))
	}

	return
}
