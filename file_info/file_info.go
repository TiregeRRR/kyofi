package fileinfo

import "io"

type FileInfo struct {
	Name      string
	Size      string
	Permision string
}

type CopyInfo struct {
	Source io.Reader
	Name   string
	Path   string
	Size   int64
}

type Copier interface {
	Next() bool
	File() (CopyInfo, error)
}
