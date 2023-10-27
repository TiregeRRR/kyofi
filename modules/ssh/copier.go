package ssh

import (
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
	"github.com/pkg/sftp"
)

type SSHCopier struct {
	sftpClient *sftp.Client

	base  string
	cnt   atomic.Int64
	paths []string

	pCh chan<- string
}

func (s *SSHCopier) File() (fileinfo.CopyInfo, error) {
	if int(s.cnt.Load()) >= len(s.paths) {
		return fileinfo.CopyInfo{}, nil
	}

	defer s.cnt.Add(1)

	file, err := s.sftpClient.Open(s.paths[s.cnt.Load()])
	if err != nil {
		return fileinfo.CopyInfo{}, err
	}

	inf, err := file.Stat()
	if err != nil {
		return fileinfo.CopyInfo{}, err
	}

	p := utils.NewProgresser(file.Name(), inf.Size(), s.pCh)
	teeFile := io.TeeReader(file, p)

	pathWithoutFile := filepath.Dir(s.paths[s.cnt.Load()])
	relativePath := strings.TrimPrefix(pathWithoutFile, s.base)

	return fileinfo.CopyInfo{
		Source: teeFile,
		Name:   inf.Name(),
		Path:   relativePath,
		Size:   inf.Size(),
	}, nil
}

func (s *SSHCopier) Next() bool {
	return int(s.cnt.Load()) < len(s.paths)
}
