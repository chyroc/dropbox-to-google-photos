package iface

import (
	"io"
)

type FileItem interface {
	Open() (io.Reader, int64, error)
	Name() string
	Size() int64
}

type FileItemSeeker interface {
	OpenSeeker() (io.ReadSeekCloser, int64, error)
	Name() string
	Size() int64
}
