package minio

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	minioClient *minio.Client

	copyBucket string
	copyPath   string
	curBucket  string
	curPath    string

	pCh chan string
}

type Opts struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SkipVerify      bool
	UseSSL          bool
}

func New(opts Opts) (*Minio, error) {
	tr := http.DefaultClient.Transport
	if opts.SkipVerify {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	cl, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(opts.AccessKeyID, opts.SecretAccessKey, ""),
		Transport: tr,
		Secure:    opts.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	if _, err := cl.HealthCheck(time.Second); err != nil {
		return nil, err
	}

	return &Minio{
		minioClient: cl,
		pCh:         make(chan string),
	}, nil
}

func (m *Minio) Open(path string) ([]fileinfo.FileInfo, error) {
	switch {
	case path == "" && m.curBucket == "":
		list, err := m.bucketList()
		if err != nil {
			return nil, err
		}

		return list, nil

	case m.curBucket == "":
		m.curBucket = path

		files, err := m.bucketFiles()
		if err != nil {
			m.curBucket = ""
			return nil, err
		}

		return files, nil
	default:
		if path != "" {
			m.curPath = path
		}

		files, err := m.bucketFiles()
		if err != nil {
			m.curBucket = ""
			return nil, err
		}

		return files, nil
	}
}

func (m *Minio) Back() ([]fileinfo.FileInfo, error) {
	switch {
	case m.curPath != "":
		f := strings.Split(m.curPath, "/")
		if len(f) == 2 {
			m.curPath = ""
			return m.bucketFiles()
		}

		m.curPath = strings.Join(f[:len(f)-1], "/")

		return m.bucketFiles()
	case m.curBucket != "":
		m.curBucket = ""
		return m.bucketList()
	}

	return nil, nil
}

func (m *Minio) Copy(name string) error {
	m.copyBucket = m.curBucket
	m.copyPath = filepath.Join(m.curPath, name)
	if strings.Contains(name, "/") {
		m.copyPath += "/"
	}

	return nil
}

func (m *Minio) PasteReader() (fileinfo.Copier, error) {
	if m.copyPath == "" {
		return nil, errors.New("nothing in buffer")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	absPart := m.copyPath

	if !strings.HasSuffix(m.copyPath, "/") {
		inf, err := m.minioClient.StatObject(ctx, m.copyBucket, m.copyPath, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("stat object %s bucket %s failed: %w", m.copyPath, m.copyBucket, err)
		}

		if inf.Size != 0 {
			return &MinioCopier{
				minioClient: m.minioClient,
				bucket:      m.copyBucket,
				base:        filepath.Dir(m.copyPath),
				paths:       []string{m.copyPath},
				pCh:         m.pCh,
			}, nil
		}

		absPart = filepath.Dir(m.copyPath) + "/"
	}

	cop := MinioCopier{
		minioClient: m.minioClient,
		bucket:      m.copyBucket,
		base:        absPart,
		pCh:         m.pCh,
	}

	for obj := range m.minioClient.ListObjects(ctx, m.curBucket, minio.ListObjectsOptions{Prefix: absPart}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("list object %s failed: %w", obj.Key, obj.Err)
		}

		if obj.Size == 0 {
			continue
		}

		cop.paths = append(cop.paths, obj.Key)
	}

	return &cop, nil
}

func (m *Minio) Paste(cop fileinfo.Copier) error {
	for cop.Next() {
		_, err := cop.File()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Minio) Delete(string) error {
	return nil
}

func (m *Minio) ProgressLogs() <-chan string {
	return m.pCh
}

func (m *Minio) Close() {
	close(m.pCh)
}

func (m *Minio) bucketList() ([]fileinfo.FileInfo, error) {
	buckets, err := m.minioClient.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}

	s := make([]fileinfo.FileInfo, len(buckets))

	for i, bucket := range buckets {
		s[i].Name = bucket.Name
		s[i].Size = "BUCKET"
		s[i].Permision = ""
	}

	return s, nil
}

func (m *Minio) bucketFiles() ([]fileinfo.FileInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	f := []fileinfo.FileInfo{}

	var err error

	for obj := range m.minioClient.ListObjects(ctx, m.curBucket, minio.ListObjectsOptions{Prefix: m.curPath}) {
		if obj.Err != nil {
			err = obj.Err
		}

		name := strings.TrimPrefix(obj.Key, m.curPath)
		size := "DIR"
		if obj.Size != 0 {
			size = utils.FormatBytes(float64(obj.Size))
		}

		f = append(f, fileinfo.FileInfo{
			Name:      name,
			Size:      size,
			Permision: obj.Owner.DisplayName,
		})
	}

	return f, err
}
