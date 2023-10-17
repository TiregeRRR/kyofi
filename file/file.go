package file

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
)

type File struct {
	curDir string
}

func New(path string) *File {
	return &File{curDir: path}
}

func (f *File) Open(path string) ([]fileinfo.FileInfo, error) {
	if f.curDir == "" {
		f.curDir = path
		return getFiles(path)
	}

	newPath := filepath.Join(f.curDir, path)
	fi, err := os.Stat(newPath)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, errors.New("not implemented")
	}

	f.curDir = newPath

	return getFiles(newPath)
}

func (f *File) Back() ([]fileinfo.FileInfo, error) {
	list := strings.Split(f.curDir, "/")
	newPath := filepath.Join(list[0 : len(list)-1]...)

	f.curDir = "/" + newPath

	return getFiles("/" + newPath)
}

func getFiles(path string) ([]fileinfo.FileInfo, error) {
	f, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	s := make([]fileinfo.FileInfo, len(f))
	for i := range f {
		s[i].Name = f[i].Name()
		inf, err := f[i].Info()
		if err != nil {
			return nil, err
		}
		s[i].Permision = inf.Mode().String()
		if !inf.IsDir() {
			s[i].Size = formatBytes(float64(inf.Size()))
		} else {
			s[i].Size = "dir"
		}
	}

	return s, nil
}

// TODO rewrite
var suffixes [5]string

func formatBytes(size float64) string {
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"

	base := math.Log(size) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := "unk"
	if int(math.Floor(base)) < 5 && int(math.Floor(base)) >= 0 {
		getSuffix = suffixes[int(math.Floor(base))]
	}
	return strconv.FormatFloat(getSize, 'f', -1, 64) + " " + string(getSuffix)
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
