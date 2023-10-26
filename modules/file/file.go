package file

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
)

type File struct {
	copyPath string
	curDir   string

	progress chan string
}

func New(path string) *File {
	return &File{
		curDir:   path,
		progress: make(chan string),
	}
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
	f.curDir = filepath.Dir(f.curDir)

	return getFiles(f.curDir)
}

func (f *File) Copy(name string) error {
	f.copyPath = filepath.Join(f.curDir, name)

	return nil
}

func (f *File) PasteReader() (fileinfo.Copier, error) {
	if f.copyPath == "" {
		return nil, errors.New("nothing in buffer")
	}

	inf, err := os.Stat(f.copyPath)
	if err != nil {
		return nil, err
	}

	if !inf.IsDir() {
		return &FileCopier{
			base:  filepath.Dir(f.copyPath),
			paths: []string{f.copyPath},
			size:  inf.Size(),
			pCh:   f.progress,
		}, nil
	}

	absPart := filepath.Dir(f.copyPath) + "/"

	cop := FileCopier{
		base: absPart,
		pCh:  f.progress,
	}
	if err := filepath.WalkDir(f.copyPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		cop.paths = append(cop.paths, path)

		inf, err := d.Info()
		if err != nil {
			return err
		}

		cop.size += inf.Size()

		return nil
	}); err != nil {
		return nil, err
	}

	return &cop, nil
}

func (f *File) Paste(cop fileinfo.Copier) error {
	for cop.Next() {
		c, err := cop.File()
		if err != nil {
			return err
		}

		if c.Path != "" {
			if err := os.MkdirAll(c.Path, 0770); err != nil {
				return err
			}
		}

		cf, err := os.Create(filepath.Join(c.Path, c.Name))
		if err != nil {
			return err
		}

		if c.Size > 1<<20 {
			go func() {
				if _, err := io.Copy(cf, c.Source); err != nil {
					panic(err)
				}

				cf.Close()
			}()
		} else {
			if _, err := io.Copy(cf, c.Source); err != nil {
				return err
			}

			cf.Close()
		}

	}

	return nil
}

func (f *File) Delete(name string) error {
	return os.RemoveAll(filepath.Join(f.curDir, name))
}

func (f *File) ProgressLogs() <-chan string {
	return f.progress
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
