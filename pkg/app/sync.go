package app

import (
	"fmt"
	"time"

	"github.com/chyroc/dropbox-to-google-photos/pkg/filetracker"
	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

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
		syncer := &sync{
			dropboxFiles:      r.dropboxFiles,
			googlePhotoClient: r.googlePhotoClient,
			fileTracker:       r.fileTracker,
			logger:            r.logger,
			Files:             make(chan iface.FileItem, 100),
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

	go func() {
		for i := 0; i < r.config.Worker; i++ {
			go func() {
				for {
					select {
					case item, ok := <-syncer.Files:
						if !ok {
							return
						}
						result := syncer.uploadFile(item)
						if result == UploadResultWait {
							r.logger.Infof("[google] upload file quote, sleep")
							time.Sleep(time.Second * 10)
						}
						if result == UploadResultWait || result == UploadResultRetry {
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

type Syncer interface {
	Run() error
}

type sync struct {
	dropboxFiles      files.Client
	googlePhotoClient *googlephotoclient.Client
	fileTracker       *filetracker.FileTracker
	logger            iface.Logger

	Files   chan iface.FileItem
	Cursor  string
	HasMore bool
}
