package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type syncer struct {
	dropboxFiles      files.Client
	googlePhotoClient *googlephotoclient.Client
	fileTracker       iface.Storer
	logger            iface.Logger

	Files   chan iface.FileItem
	Cursor  string
	HasMore bool
}

func (r *App) Sync(ignoreCursor bool) error {
	r.logger.Infof("[sync] start sync, path: '%s'", r.config.Dropbox.RootDir)

	syncer := &syncer{
		dropboxFiles:      r.dropboxFiles,
		googlePhotoClient: r.googlePhotoClient,
		fileTracker:       r.fileTracker,
		logger:            r.logger,
		Files:             make(chan iface.FileItem, 100),
	}
	var entities []files.IsMetadata

	cursor := syncer.getCursor()
	if !ignoreCursor && len(cursor) > 0 {
		syncer.updateCursor(cursor, true)
		r.logger.Infof("[sync] load cursor: '%s'", cursor)
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
				r.logger.Debugf("[sync] append file: '%s', hash: '%s'", fi.Name(), fi.(*dropboxFileItem).hash)
				syncer.Files <- fi
			}
		}
		for syncer.HasMore {
			err := syncer.loadDropboxImages()
			if err != nil {
				r.logger.Errorf("[sync] load images fail: %s", err)
				time.Sleep(time.Second * 3)
			}
		}
	}()

	r.logger.Infof("[sync] wait for channel")
	time.Sleep(time.Second * 3)

	reactLimit := int32(0)
	wait := new(sync.WaitGroup)
	for i := 0; i < r.config.Worker; i++ {
		wait.Add(1)
		go func() {
			defer func() { wait.Done() }()
			for {
				select {
				default:
					if syncer.HasMore {
						time.Sleep(time.Second * 3)
						continue
					}
					return
				case item := <-syncer.Files:
					switch syncer.uploadFile(item) {
					case UploadResultReactDayLimit:
						atomic.AddInt32(&reactLimit, 1)
						r.logger.Infof("[sync] limit per day, return and stop")
						return
					case UploadResultWait:
						r.logger.Infof("[sync] upload file quote, sleep")
						time.Sleep(time.Second * 10)
						go func() { syncer.Files <- item }()
					case UploadResultRetry:
						go func() { syncer.Files <- item }()
					}
				}
			}
		}()
	}
	wait.Wait()

	if atomic.LoadInt32(&reactLimit) > 0 {
		r.logger.Infof("[sync] react limit, stop")
	} else {
		r.logger.Infof("[sync] done")
	}

	return nil
}
