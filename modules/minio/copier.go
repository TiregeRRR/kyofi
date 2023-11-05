package minio

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
	"github.com/minio/minio-go/v7"
)

type MinioCopier struct {
	minioClient *minio.Client

	bucket string
	base   string
	cnt    atomic.Int64
	paths  []string

	pCh chan<- string
}

func (m *MinioCopier) File() (fileinfo.CopyInfo, error) {
	cnt := m.cnt.Load()

	if int(cnt) >= len(m.paths) {
		return fileinfo.CopyInfo{}, nil
	}

	defer m.cnt.Add(1)

	m.pCh <- m.paths[cnt]

	file, err := m.minioClient.GetObject(context.Background(), m.bucket, m.paths[cnt], minio.GetObjectOptions{})
	if err != nil {
		return fileinfo.CopyInfo{}, fmt.Errorf("get object %s from bucket %s failed: %w", m.paths[cnt], m.bucket, err)
	}

	inf, err := file.Stat()
	if err != nil {
		return fileinfo.CopyInfo{}, fmt.Errorf("stat object %s from bucket %s failed: %w", m.paths[cnt], m.bucket, err)
	}

	m.pCh <- inf.Key

	p := utils.NewProgresser(inf.Key, inf.Size, m.pCh)
	teeFile := io.TeeReader(file, p)

	path := filepath.Dir(m.paths[cnt])
	relativePath := strings.TrimSuffix(path, m.base)

	return fileinfo.CopyInfo{
		Source: teeFile,
		Name:   filepath.Base(inf.Key),
		Path:   relativePath,
		Size:   inf.Size,
	}, nil
}

func (m *MinioCopier) Next() bool {
	return int(m.cnt.Load()) < len(m.paths)
}
