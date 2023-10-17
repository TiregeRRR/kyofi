package minio

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	minioClient *minio.Client
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

	if _, err := cl.ListBuckets(context.Background()); err != nil {
		return nil, err
	}

	return &Minio{
		minioClient: cl,
	}, nil
}

func (m *Minio) Open(path string) ([]fileinfo.FileInfo, error) {
	switch path {
	case "":

		obj, err := m.minioClient.GetObject(context.Background(), "datasets", "PROJECT/OI/ROMEN_test/event_pubs.parquet", minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		list, err := m.bucketList()
		if err != nil {
			return nil, err
		}

		st, err := obj.Stat()
		if err != nil {
			return nil, err
		}

		list = append(list, fileinfo.FileInfo{
			Name:      st.Key,
			Size:      utils.FormatBytes(float64(st.Size)),
			Permision: "",
		})

		return list, nil
	default:
		return nil, nil
	}
}

func (m *Minio) Back() ([]fileinfo.FileInfo, error) {
	return nil, nil
}

func (m *Minio) Copy(name string) error {
	return nil
}

func (m *Minio) PasteReader() (io.Reader, string, error) {
	return nil, "", nil
}

func (m *Minio) Paste(string, io.Reader) error {
	return nil
}

func (m *Minio) Delete(string) error {
	return nil
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
