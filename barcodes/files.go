package barcodes

import (
	"fastq_demultiplexer/chan_io"
	"fastq_demultiplexer/fastq"
	"fastq_demultiplexer/structs"
	"os"
	"sync"
)

type BarcodeFiles map[string]*BarcodeFileGroup

type BarcodeFileGroup struct {
	Name string

	Data *FileNameTemplateData

	Files map[string]*BarcodeFile
}

type BarcodeFile struct {
	file      *os.File
	innerFile *os.File

	bytes uint64

	appOptions  *structs.AppOptions
	fileOptions *structs.FileOptions

	ChanIO *chan_io.IOStringLines

	closeHandler func() error
}

func (g *BarcodeFileGroup) Init(options *structs.AppOptions, schema *[]string, wg *sync.WaitGroup) (err error) {
	template := FilenameTemplate{templateString: options.FilenameTemplate}

	err = template.Init()

	if err != nil {
		return
	}

	g.Files = make(map[string]*BarcodeFile)

	var path string

	for _, schemaType := range *schema {
		g.Data.ReadType = schemaType
		path, err = template.Parse(*g.Data)

		if err != nil {
			return
		}

		g.Files[schemaType], err = (&BarcodeFile{}).Init(options, g.Data.SampleName, path, wg)

		if err != nil {
			return
		}
	}

	return
}

func (f *BarcodeFile) Init(appOptions *structs.AppOptions, name, filename string, wg *sync.WaitGroup) (file *BarcodeFile, err error) {
	fileOptions, err := structs.BuildFileOptions(name, filename, appOptions.OutputDirectoryPath, appOptions.Debug)

	f.ChanIO = chan_io.NewIOStringLines(appOptions.BlockSize)

	err = fastq.WriteFile(appOptions, *fileOptions, f.ChanIO, wg)

	file = f

	return
}

func EnsureFileChannelsClose(schema *[]string, groupFiles *BarcodeFiles) {
	var wg sync.WaitGroup

	for _, group := range *groupFiles {
		wg.Add(len(group.Files))

		go func(g *BarcodeFileGroup) {
			for _, schemaType := range *schema {
				go g.Files[schemaType].ChanIO.EnsureWriteWithClose(&wg)
			}
		}(group)
	}

	wg.Wait()
}
