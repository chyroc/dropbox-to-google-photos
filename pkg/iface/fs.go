package iface

import (
	"io"
)

type FileItem interface {
	OpenSeeker() (io.ReadSeekCloser, int64, error)
	Name() string
	Size() int64
}
