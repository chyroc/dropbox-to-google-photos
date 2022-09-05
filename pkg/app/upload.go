package app

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
)

func (r *App) UploadPath(path string) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fs := googlephotoclient.NewFileItem(filepath.Base(path), int64(len(bs)), bytes.NewReader(bs))

	uploadToken, err := r.googlePhotoClient.UploadFile(context.Background(), fs)
	if err != nil {
		return err
	}
	res, err := r.googlePhotoClient.UploadFileToLibrary(context.Background(), uploadToken)
	if err != nil {
		r.logger.Errorf("[sync] upload fail: '%s': %s", fs.Name(), err)
		return err
	} else {
		r.logger.Infof("[sync] upload success, id: '%s', name: '%s'", res.ID, res.Filename)
	}
	_ = res
	return nil
}
