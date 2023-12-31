package file

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
)

type FileCopier struct {
	base  string
	cnt   atomic.Int64
	paths []string

	pCh chan<- string
}

func (f *FileCopier) File() (fileinfo.CopyInfo, error) {
	if int(f.cnt.Load()) >= len(f.paths) {
		return fileinfo.CopyInfo{}, nil
	}

	defer f.cnt.Add(1)

	file, err := os.Open(f.paths[f.cnt.Load()])
	if err != nil {
		return fileinfo.CopyInfo{}, err
	}

	inf, err := file.Stat()
	if err != nil {
		return fileinfo.CopyInfo{}, err
	}

	p := utils.NewProgresser(inf.Name(), inf.Size(), f.pCh)
	teeFile := io.TeeReader(file, p)

	pathWithoutFile := filepath.Dir(f.paths[f.cnt.Load()])
	relativePath := strings.TrimPrefix(pathWithoutFile, f.base)

	return fileinfo.CopyInfo{
		Source: teeFile,
		Name:   inf.Name(),
		Path:   relativePath,
		Size:   inf.Size(),
	}, nil
}

func (f *FileCopier) Next() bool {
	return int(f.cnt.Load()) < len(f.paths)
}
