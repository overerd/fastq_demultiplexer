package misc

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
)

const reCompressionQuery = "\\.(?:[gG][zZ])$"
const reFilenameQuery = "[\\/]?([^\\/]+)$"

var reCompression *regexp.Regexp
var reFilename *regexp.Regexp

var rPathSeparator *regexp.Regexp
var rWrongPathSeparator *regexp.Regexp

func InitMisc() (err error) {
	err = InitArchSpecifics()

	if err != nil {
		return
	}

	err = InitRegex()

	if err != nil {
		return
	}

	return
}

func InitArchSpecifics() (err error) {
	rPathSeparator, err = regexp.Compile(fmt.Sprintf("%s%s+", PathSeparator, PathSeparator))
	rWrongPathSeparator, err = regexp.Compile(fmt.Sprintf("%s+", WrongPathSeparator))

	return
}

func InitRegex() (err error) {
	reCompression, err = regexp.Compile(reCompressionQuery)

	if err != nil {
		return
	}

	reFilename, err = regexp.Compile(reFilenameQuery)

	return
}

func ExtractMapStringKeys(m *map[string]interface{}) (res []string) {
	for key := range *m {
		res = append(res, key)
	}

	sort.Strings(res)

	return
}

func ExtractFilename(path string) (filename string, err error) {
	filename = reFilename.FindString(path)

	if filename == "" {
		err = errors.New(fmt.Sprintf("extracted empty filename"))
	}

	return
}

func CheckFilenameForCompression(path string) bool {
	return reCompression.MatchString(path)
}

func ConcatenatePath(path, mod string) string {
	path = (*rPathSeparator).ReplaceAllString((*rWrongPathSeparator).ReplaceAllString(path, PathSeparator), PathSeparator)

	return fmt.Sprintf("%s/%s", path, mod)
}
