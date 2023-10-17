package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
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
			s[i].Size = utils.FormatBytes(float64(inf.Size()))
		} else {
			s[i].Size = "DIR"
		}
	}

	return s, nil
}