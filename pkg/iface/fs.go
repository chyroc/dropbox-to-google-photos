package iface

import (
	"io"
)

type FileItem interface {
	OpenSeeker() (io.ReadSeeker, int64, error)
	Name() string
	Size() int64
}
