package barcodes

import (
	"bufio"
	"fastq_demultiplexer/structs"
	"fmt"
	"os"
	"strings"
	"sync"
)

type BarcodeMap map[string]string

func LoadTargets(options *structs.AppOptions) (targets map[string]bool, err error) {
	targets = make(map[string]bool)

	if options.TargetsPath == "" {
		return
	}

	file, err := os.OpenFile(options.TargetsPath, os.O_RDONLY, 0555)

	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanWords)

	buffer := make([]byte, options.BufferSize)

	scanner.Buffer(buffer, int(options.BufferSize)*10)

	for scanner.Scan() {
		target := scanner.Text()

		if target == "" {
			continue
		}

		targets[target] = true
	}

	return
}

func Load(options *structs.AppOptions, targets map[string]bool, schema *[]string, wg *sync.WaitGroup) (n int, barcodes BarcodeMap, files BarcodeFiles, err error) {
	barcodes = make(BarcodeMap)
	files = make(BarcodeFiles)

	n = 0

	file, err := os.OpenFile(options.TablePath, os.O_RDONLY, 0555)

	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanWords)

	buffer := make([]byte, options.BufferSize)

	scanner.Buffer(buffer, int(options.BufferSize)*10)

	var sampleNumber uint = 0

	for scanner.Scan() {
		line := scanner.Text()

		columns := strings.Split(line, ",")

		key := columns[0]

		keyItems := strings.Split(key, "-")

		key = strings.Join(keyItems[:3], "-")

		for _, barcode := range columns[1:] {
			// preparing os.File instance for each target
			if _, ok := files[key]; !ok {
				if len(targets) > 0 {
					if _, acquired := targets[key]; !acquired {
						continue
					}
				}

				sampleNumber++

				fmt.Printf(" [!] adding target '%s' #%d\n", key, sampleNumber)

				barcodeFileGroup := &BarcodeFileGroup{
					Name: key,
					Data: &FileNameTemplateData{
						SampleNumber: sampleNumber,
						SampleName:   key,
						LaneNumber:   options.LaneNumber,
					},
				}

				err = barcodeFileGroup.Init(options, schema, wg)

				if err != nil {
					return
				}

				files[key] = barcodeFileGroup

				wg.Add(len(barcodeFileGroup.Files))
				n++
			}

			barcode = strings.ToUpper(barcode)

			// inflating barcodes with single nucleotide differences
			for i := range barcode {
				for _, char := range "ATGC" {
					hash := barcode[:i] + string(char) + barcode[i+1:]

					barcodes[hash] = key
				}
			}
		}
	}

	err = scanner.Err()

	return
}
