package app

import (
	"context"
	"fmt"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func (r *sync) uploadFile(item iface.FileItem) UploadResult {
	var err error

	// check if file is already uploaded
	if r.checkFileExist(item) {
		r.logger.Infof("[sync] file exist: '%s', skip", item.Name())
		return ""
	}

	r.logger.Infof("[sync] uploading file: '%s', size: %s", item.Name(), humanSize(item.Size()))

	// check if upload token exist
	uploadToken := r.getUploadToken(item)
	if uploadToken == "" {
		uploadToken, err = r.googlePhotoClient.UploadFile(context.Background(), item)
		if err != nil {
			result := wrapGoogleError(err)
			if result == UploadResultWait || result == UploadResultReactDayLimit {
				return result
			}
			r.logger.Errorf("[sync] upload token fail: '%s': %s", item.Name(), err)
			return result
		}
		r.setUploadToken(item, uploadToken)
	}

	media, err := r.googlePhotoClient.UploadFileToLibrary(context.Background(), uploadToken)
	if err != nil {
		result := wrapGoogleError(err)
		if result == UploadResultWait || result == UploadResultReactDayLimit {
			return result
		}
		r.logger.Errorf("[sync] add library fail: '%s': %s", item.Name(), err)
		return result
	}

	r.setFileExist(item)

	if item.Name() != media.Filename {
		r.logger.Infof("[sync] upload success: '%s', name: '%s'", item.Name(), media.Filename)
	} else {
		r.logger.Infof("[sync] upload success: '%s'", item.Name())
	}

	return ""
}

func (r *sync) itemToExistKey(item iface.FileItem) string {
	return "dropbox.hash:" + item.(*dropboxFileItem).hash
}

func (r *sync) itemToUploadTokenKey(item iface.FileItem) string {
	return "dropbox-to-google.upload_token:" + item.(*dropboxFileItem).hash
}

func (r *sync) checkFileExist(item iface.FileItem) bool {
	return len(r.fileTracker.Get(r.itemToExistKey(item))) > 0
}

func (r *sync) setFileExist(item iface.FileItem) {
	r.fileTracker.Set(r.itemToExistKey(item), item.(*dropboxFileItem).hash)
}

func (r *sync) getUploadToken(item iface.FileItem) string {
	return r.fileTracker.Get(r.itemToUploadTokenKey(item))
}

func (r *sync) setUploadToken(item iface.FileItem, uploadToken string) {
	r.fileTracker.Set(r.itemToUploadTokenKey(item), uploadToken)
}

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f kB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
	}
}
