package iface

import (
	"io"
)

type FileItem interface {
	Open() (io.Reader, int64, error)
	Name() string
	Size() int64
}
