package app

import (
	"context"
	"fmt"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func (r *sync) uploadFile(item iface.FileItem) UploadResult {
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
			// 200M: 1024*1024*200
			if item.Size() > 1024*1024*200 {
				r.logger.Debugf("[sync] upload file with part: '%s', size: %s", item.Name(), humanSize(item.Size()))
				uploadToken, err = r.googlePhotoClient.UploadFilePart(context.Background(), item)
			} else {
				uploadToken, err = r.googlePhotoClient.UploadFile(context.Background(), item)
			}
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
		if result == UpdateResultSkip {
			r.setFileSkip(item)
			return result
		}
		if result == UploadResultWait || result == UploadResultReactDayLimit || result == UploadResultRetry {
			return result
		}
		r.logger.Errorf("[sync] upload fail: '%s': %s", item.Name(), err)
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
