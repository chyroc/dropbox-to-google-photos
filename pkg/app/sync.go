package app

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chyroc/dropbox-to-google-photos/pkg/filetracker"
	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type sync struct {
	dropboxFiles      files.Client
	googlePhotoClient *googlephotoclient.Client
	fileTracker       *filetracker.FileTracker
	logger            iface.Logger

	Files   chan iface.FileItem
	Cursor  string
	HasMore bool
}

func (r *App) Sync() error {
	r.logger.Infof("start sync, path: '%s'", r.config.Dropbox.RootDir)

	syncer := &sync{
		dropboxFiles:      r.dropboxFiles,
		googlePhotoClient: r.googlePhotoClient,
		fileTracker:       r.fileTracker,
		logger:            r.logger,
		Files:             make(chan iface.FileItem, 100),
	}
	var entities []files.IsMetadata

	if cursor := syncer.getCursor(); len(cursor) > 0 {
		syncer.updateCursor(cursor, true)
		r.logger.Infof("[dropbox] load cursor: '%s'", cursor)
	} else {
		res, err := r.dropboxFiles.ListFolder(&files.ListFolderArg{
			Path:      r.config.Dropbox.RootDir,
			Recursive: true,
			Limit:     1000,
		})
		if err != nil {
			return fmt.Errorf("dropbox list folder fail: %w", err)
		}
		syncer.updateCursor(res.Cursor, res.HasMore)
		entities = res.Entries
	}

	// load dropbox images
	go func() {
		for _, v := range entities {
			if fi := syncer.dropboxMetadataToFileItem(v); fi != nil {
				r.logger.Infof("[dropbox] append file: '%s', hash: '%s'", fi.Name(), fi.(*dropboxFileItem).hash)
				syncer.Files <- fi
			}
		}
		for syncer.HasMore {
			err := syncer.loadDropboxImages()
			if err != nil {
				r.logger.Errorf("[dropbox] load images fail: %s", err)
				time.Sleep(time.Second * 3)
			}
		}
	}()

	worker := int32(r.config.Worker)
	go func() {
		for i := 0; i < r.config.Worker; i++ {
			go func() {
				defer func() {
					atomic.AddInt32(&worker, -1)
				}()
				for {
					select {
					case item, ok := <-syncer.Files:
						if !ok {
							return
						}
						switch syncer.uploadFile(item) {
						case UploadResultReturn:
							r.logger.Infof("[dropbox] limit per day, return and stop")
							return
						case UploadResultWait:
							r.logger.Infof("[google] upload file quote, sleep")
							time.Sleep(time.Second * 10)
							go func() { syncer.Files <- item }()
						case UploadResultRetry:
							go func() { syncer.Files <- item }()
						}
					}
				}
			}()
		}
	}()

	x := time.NewTicker(time.Second)
	time.Sleep(time.Second * 5)
	for {
		if atomic.LoadInt32(&worker) == 0 {
			r.logger.Infof("sync done")
			break
		}
		select {
		case <-x.C:
			if !syncer.HasMore {
				r.logger.Infof("sync done")
				return nil
			}
		}
	}
	return nil
}
