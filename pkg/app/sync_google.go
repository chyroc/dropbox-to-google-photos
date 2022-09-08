package app

import (
	"context"
	"fmt"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func (r *syncer) uploadFile(item iface.FileItem) UploadResult {
	var err error

	// check if file is already uploaded
	if r.checkFileExist(item) {
		r.logger.Infof("[sync] skip this file: '%s', exist", item.Name())
		return ""
	}

	r.logger.Infof("[sync] uploading file: '%s', size: %s", item.Name(), humanSize(item.Size()))

	var media googlephotoclient.MediaItem

	{
		// 检查 token 缓存
		uploadToken := r.getUploadToken(item)

		// 如果没有，则上传图片，获取 token
		if uploadToken == "" {
			uploadToken, err = r.googlePhotoClient.UploadFilePart(context.Background(), item)
			if uploadToken != "" {
				r.setUploadToken(item, uploadToken)
			}
		}

		// 如果获取 token 成功，则添加到相册
		if err == nil {
			media, err = r.googlePhotoClient.UploadFileToLibrary(context.Background(), uploadToken)
		}
	}

	// 如果 有 err（可能是上传图片产生的，有可能是添加到相册产生的）
	if err != nil {
		result := wrapGoogleError(err)
		switch result {
		case UpdateResultSkip:
			r.logger.Debugf("[sync] skip this file: '%s', exist", item.Name())
		case UploadResultWaitAndRetry:
			r.logger.Debugf("[sync] upload file quote")
		case UploadResultRetry:
			r.logger.Debugf("[sync] upload file retry")
		case UploadResultReactDayLimit:
			r.logger.Debugf("[sync] upload file react day limit")
		case UpdateResultError:
			r.logger.Errorf("[sync] upload fail: '%s': %s", item.Name(), err)
		default:
			// do nothing
		}
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
