package main

import (
	"compress/gzip"
	"fastq_demultiplexer/barcodes"
	"fastq_demultiplexer/chan_io"
	"fastq_demultiplexer/fastq"
	"fastq_demultiplexer/misc"
	"fastq_demultiplexer/structs"
	"fastq_demultiplexer/transform"
	"fastq_demultiplexer/transform/transform_strategies"
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	version = "0.1.0"
	build   = "src"
)

const AppTitle = "fastq converter/demultiplexer"
const AppDescription = "Transforms fastq files from MGI to Illumina format."

func invokeError(err error) {
	if err == nil {
		return
	}

	println("\n")

	log.Fatalln(err.Error())
}

func main() {
	appString := fmt.Sprintf("%s v%s (build %s)", AppTitle, version, build)

	fmt.Println(fmt.Sprintf("Starting %s...\n", appString))

	options, r1ReaderOptions, r2ReaderOptions, err := setup()

	if err != nil {
		invokeError(err)
	}

	invokeError(run(options, r1ReaderOptions, r2ReaderOptions))

	println()
}

func setup() (options structs.AppOptions, r1ReaderOptions, r2ReaderOptions structs.FileOptions, err error) {
	parser := argparse.NewParser(os.Args[0], AppDescription)

	transform_strategies.InitStrategies()

	_, supportedStrategies, err := transform_strategies.GetAllStrategies()

	if err != nil {
		invokeError(err)
	}

	r1Path := parser.String("1", "r1", &argparse.Options{
		Required: true,
		Help:     "path to a R1 fastq-file",
	})

	r2Path := parser.String("2", "r2", &argparse.Options{
		Required: true,
		Help:     "path to a R2 fastq-file",
	})

	tablePath := parser.String("c", "csv-file", &argparse.Options{
		Required: true,
		Help:     "path to barcode csv-file",
	})

	tableSeparatorPath := parser.String("s", "csv-separator", &argparse.Options{
		Required: false,
		Default:  ",",
		Help:     "barcode csv-file separator",
	})

	outputDirectory := parser.String("o", "output-directory", &argparse.Options{
		Required: true,
		Help:     "path to output directory",
	})

	targets := parser.String("", "targets-file", &argparse.Options{
		Required: false,
		Help:     "path to file with targets (if null, would select all possible indexes from barcodes file)",
	})

	transformStrategy := parser.String("", "transform-strategy", &argparse.Options{
		Required: false,
		Default:  "base",
		Help: fmt.Sprintf(
			"strategy of how to transform fastq data (supported strategies: %s)",
			strings.Join(supportedStrategies, ", "),
		),
	})

	filenameTemplate := parser.String("", "filename-template", &argparse.Options{
		Required: false,
		Default:  barcodes.DefaultFilenameTemplate,
		Help:     fmt.Sprintf("filename template (default: '%s')", barcodes.DefaultFilenameTemplate),
	})

	laneNumber := parser.Int("", "lane-number", &argparse.Options{
		Required: false,
		Default:  1,
		Help:     "lane number for selected fastq pair",
	})

	bufferSize := parser.Int("", "buffer-size", &argparse.Options{
		Required: false,
		Default:  10 * 1024 * 1024,
		Help:     "I/O buffer size",
	})

	blockSize := parser.Int("", "block-size", &argparse.Options{
		Required: false,
		Default:  4 * 2 * 1024,
		Help:     "I/O block size",
	})

	compressionLevel := parser.Int("", "compression-level", &argparse.Options{
		Required: false,
		Default:  gzip.BestSpeed,
		Help:     "compression level [1, 9]",
	})

	debugFlag := parser.Flag("", "debug", &argparse.Options{
		Help: "enable debug messages",
	})

	invokeError(parser.Parse(os.Args))

	err = misc.InitMisc()

	if err != nil {
		return
	}

	options = structs.AppOptions{
		R1Path: *r1Path,
		R2Path: *r2Path,

		TargetsPath: *targets,

		TablePath:      *tablePath,
		TableSeparator: *tableSeparatorPath,

		FilenameTemplate: *filenameTemplate,
		LaneNumber:       uint(*laneNumber),

		OutputDirectoryPath: *outputDirectory,

		TransformStrategy: *transformStrategy,

		BufferSize: uint(*bufferSize),
		BlockSize:  uint(*blockSize),

		Debug: *debugFlag,
	}

	if options.Debug {
		fmt.Printf("%s\n\n", time.Now().String())
		fmt.Println(fmt.Sprintf("%+v", options))
		fmt.Println(fmt.Sprintf("%+v", r1ReaderOptions))
		fmt.Println(fmt.Sprintf("%+v", r2ReaderOptions))
	}

	r1FileName, err := misc.ExtractFilename(options.R1Path)

	if err != nil {
		return
	}

	r2FileName, err := misc.ExtractFilename(options.R2Path)

	if err != nil {
		return
	}

	r1ReaderOptions = structs.FileOptions{
		Path:             options.R1Path,
		Filename:         r1FileName,
		Debug:            *debugFlag,
		UseCompression:   misc.CheckFilenameForCompression(options.R1Path),
		CompressionLevel: *compressionLevel,
	}

	r2ReaderOptions = structs.FileOptions{
		Path:             options.R2Path,
		Filename:         r2FileName,
		Debug:            *debugFlag,
		UseCompression:   misc.CheckFilenameForCompression(options.R2Path),
		CompressionLevel: *compressionLevel,
	}

	return
}

func run(options structs.AppOptions, r1ReaderOptions, r2ReaderOptions structs.FileOptions) (err error) {
	fmt.Printf("Reading barcode table from file '%s'...\n\n", options.TablePath)

	transformer, err := transform_strategies.GetStrategy(options.TransformStrategy)

	if err != nil {
		return
	}

	schema := transformer.GetSchema()

	targets, err := barcodes.LoadTargets(&options)

	if err != nil {
		return
	}

	var outputWaitGroup sync.WaitGroup

	_, barcodesMap, barcodeFiles, err := barcodes.Load(&options, targets, &schema, &outputWaitGroup)

	if err != nil {
		return
	}

	fmt.Printf("\nBeginning transformation...\n\n")

	r1Input := chan_io.NewIOStringLines(options.BlockSize)
	r2Input := chan_io.NewIOStringLines(options.BlockSize)

	invokeError(fastq.ReadFile(options, r1ReaderOptions, r1Input))
	invokeError(fastq.ReadFile(options, r2ReaderOptions, r2Input))

	done := make(chan bool, 0)

	transform.TransformData(
		options,

		&barcodesMap,
		&barcodeFiles,

		r1Input,
		r2Input,

		done,
	)

	fmt.Println("")

	<-done

	barcodes.EnsureFileChannelsClose(&schema, &barcodeFiles)

	outputWaitGroup.Wait()

	fmt.Println("\n\nAll is done!")

	r1Input.Close()
	r2Input.Close()

	return
}
