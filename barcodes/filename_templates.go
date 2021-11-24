package barcodes

import (
	"bytes"
	"html/template"
)

const DefaultFilenameTemplate = "{{.SampleName}}_S{{.SampleNumber}}_L00{{.LaneNumber}}_{{.ReadType}}_001.fastq.gz"

type FileNameTemplateData struct {
	SampleNumber uint   `json:"sample_number"`
	SampleName   string `json:"sample_name"`
	LaneNumber   uint   `json:"lane_number"`
	ReadType     string `json:"read_type"`
}

type FilenameTemplate struct {
	templateString string

	template *template.Template
}

func (t *FilenameTemplate) Init() (err error) {
	t.template, err = template.New("filename").Parse(t.templateString)

	return
}

func (t *FilenameTemplate) Parse(data FileNameTemplateData) (res string, err error) {
	writer := bytes.NewBufferString("")

	err = t.template.Execute(writer, data)

	if err != nil {
		return
	}

	res = writer.String()

	return
}
