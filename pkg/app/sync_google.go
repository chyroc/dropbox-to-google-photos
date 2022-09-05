package app

import (
	"context"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func (r *sync) uploadFile(item iface.FileItem) UploadResult {
	r.logger.Infof("[sync] upload file: '%s'", item.Name())

	contentHash := item.(*dropboxFileItem).hash
	// check if file is already uploaded
	value := r.fileTracker.Get("dropbox.hash:" + contentHash)
	if len(value) > 0 {
		r.logger.Infof("[google] file exist: '%s', skip", item.Name())
		return ""
	}

	media, err := r.googlePhotoClient.UploadFileToLibrary(context.Background(), item)
	if err != nil {
		result := wrapGoogleError(err)
		if result == UploadResultWait || result == UploadResultReturn {
			return result
		}
		r.logger.Errorf("[sync] upload fail: '%s': %s", item.Name(), err)
		return result
	}

	r.fileTracker.Set("dropbox.hash:"+contentHash, contentHash)

	r.logger.Infof("[sync] upload success: '%s': id: '%s', name: '%s'", item.Name(), media.ID, media.Filename)

	return ""
}
