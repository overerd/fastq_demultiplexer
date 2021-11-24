package structs

import (
	"fastq_demultiplexer/misc"
	"os"
)

type FileOptions struct {
	Path     string
	Filename string
	Debug    bool

	UseCompression   bool
	CompressionLevel int
}

func BuildFileOptions(sampleName, filename, directory string, debug bool) (options *FileOptions, err error) {
	sampleDirectory := misc.ConcatenatePath(directory, sampleName)

	err = os.MkdirAll(sampleDirectory, 0770)

	if err != nil {
		return
	}

	options = &FileOptions{
		Path:           misc.ConcatenatePath(sampleDirectory, filename),
		Filename:       filename,
		Debug:          debug,
		UseCompression: misc.CheckFilenameForCompression(filename),
	}

	return
}
