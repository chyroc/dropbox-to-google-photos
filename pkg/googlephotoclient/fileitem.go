package googlephotoclient

import (
	"io"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func NewFileItem(name string, size int64, data io.ReadSeeker) iface.FileItem {
	return &fileItemImpl{
		name: name,
		size: size,
		data: data,
	}
}

type fileItemImpl struct {
	name string
	size int64
	data io.ReadSeeker
}

func (r *fileItemImpl) OpenSeeker() (io.ReadSeeker, int64, error) {
	return r.data, r.size, nil
}

func (r *fileItemImpl) Name() string {
	return r.name
}

func (r *fileItemImpl) Size() int64 {
	return r.size
}
