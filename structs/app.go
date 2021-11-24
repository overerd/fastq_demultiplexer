package structs

type AppOptions struct {
	R1Path, R2Path, IPath              string
	R1Filename, R2Filename, I1Filename string

	TargetsPath string

	TablePath      string
	TableSeparator string

	FilenameTemplate string
	LaneNumber       uint

	OutputDirectoryPath string

	TransformStrategy string

	BufferSize uint
	BlockSize  uint
	Threads    uint16

	Debug bool
}
