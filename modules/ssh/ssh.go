package ssh

import (
	"errors"
	"io"
	"path/filepath"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
)

type SSH struct {
	copyPath string
	curDir   string

	progress chan string

	client     *goph.Client
	sftpClient *sftp.Client
}

type Opts struct {
	User     string
	Addr     string
	Password string
}

func New(opts Opts) (*SSH, error) {
	auth := goph.Password(opts.Password)

	c, err := goph.New(opts.User, opts.Addr, auth)
	if err != nil {
		return nil, err
	}

	sftp, err := c.NewSftp()
	if err != nil {
		return nil, err
	}

	curDir, err := sftp.Getwd()
	if err != nil {
		return nil, err
	}

	return &SSH{
		curDir: curDir,

		client:     c,
		sftpClient: sftp,
		progress:   make(chan string),
	}, nil
}

func (s *SSH) Open(path string) ([]fileinfo.FileInfo, error) {
	if path == "" {
		return s.getFiles(s.curDir)
	}

	newPath := filepath.Join(s.curDir, path)
	fi, err := s.sftpClient.Stat(newPath)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, errors.New("not implemented")
	}

	s.curDir = newPath

	return s.getFiles(s.curDir)
}

func (s *SSH) Back() ([]fileinfo.FileInfo, error) {
	s.curDir = filepath.Dir(s.curDir)

	return s.getFiles(s.curDir)
}

func (s *SSH) Copy(name string) error {
	s.copyPath = filepath.Join(s.curDir, name)

	return nil
}

func (s *SSH) PasteReader() (fileinfo.Copier, error) {
	if s.copyPath == "" {
		return nil, errors.New("nothing in buffer")
	}

	inf, err := s.sftpClient.Stat(s.copyPath)
	if err != nil {
		return nil, err
	}

	if !inf.IsDir() {
		return &SSHCopier{
			sftpClient: s.sftpClient,
			base:       filepath.Dir(s.copyPath),
			paths:      []string{s.copyPath},
			pCh:        s.progress,
		}, nil
	}

	absPart := filepath.Dir(s.copyPath) + "/"

	cop := SSHCopier{
		sftpClient: s.sftpClient,
		base:       absPart,
		pCh:        s.progress,
	}

	w := s.sftpClient.Walk(s.copyPath)
	for w.Step() {
		if w.Err() != nil {
			return nil, err
		}

		if w.Stat().IsDir() {
			continue
		}

		cop.paths = append(cop.paths, w.Path())
	}

	return &cop, nil
}

func (s *SSH) Paste(cop fileinfo.Copier) error {
	for cop.Next() {
		c, err := cop.File()
		if err != nil {
			return err
		}

		if c.Path != "" {
			if err := s.sftpClient.MkdirAll(c.Path); err != nil {
				return err
			}
		}

		cf, err := s.sftpClient.Create(filepath.Join(c.Path, c.Name))
		if err != nil {
			return err
		}

		if _, err := io.Copy(cf, c.Source); err != nil {
			return err
		}

		if err := cf.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (s *SSH) Delete(name string) error {
	path := filepath.Join(s.curDir, name)
	inf, err := s.sftpClient.Stat(path)
	if err != nil {
		return err
	}

	if inf.IsDir() {
		if err := s.removeAll(path); err != nil {
			return err
		}
	} else {
		if err := s.sftpClient.Remove(path); err != nil {
			return err
		}
	}

	return nil
}

// TODO from sftp 1.36, wait or fork
func (s *SSH) removeAll(path string) error {
	// Get the file/directory information
	fi, err := s.sftpClient.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		// Delete files recursively in the directory
		files, err := s.sftpClient.ReadDir(path)
		if err != nil {
			return err
		}

		for _, file := range files {
			if file.IsDir() {
				// Recursively delete subdirectories
				err = s.removeAll(path + "/" + file.Name())
				if err != nil {
					return err
				}
			} else {
				// Delete individual files
				err = s.sftpClient.Remove(path + "/" + file.Name())
				if err != nil {
					return err
				}
			}
		}

	}

	return s.sftpClient.Remove(path)
}

func (s *SSH) ProgressLogs() <-chan string {
	return s.progress
}

func (s *SSH) Close() {
	close(s.progress)
}

func (s *SSH) getFiles(path string) ([]fileinfo.FileInfo, error) {
	pathInfo, err := s.sftpClient.Stat(path)
	if err != nil {
		return nil, err
	}

	if !pathInfo.IsDir() {
		return nil, errors.New("not implemented")
	}

	f, err := s.sftpClient.ReadDir(path)
	if err != nil {
		return nil, err
	}

	fi := make([]fileinfo.FileInfo, len(f))
	for i := range f {
		fi[i].Name = f[i].Name()
		fi[i].Permision = f[i].Mode().String()
		if !f[i].IsDir() {
			fi[i].Size = utils.FormatBytes(float64(f[i].Size()))
		} else {
			fi[i].Size = "DIR"
		}
	}

	return fi, nil
}
