package app

import (
	"github.com/chyroc/dropbox-to-google-photos/pkg/filetracker"
	"github.com/chyroc/dropbox-to-google-photos/pkg/tokenmanager"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func (r *App) Start() error {
	err := r.loadConfig()
	if err != nil {
		return err
	}

	fr, err := tokenmanager.NewFileRepository(r.workDir)
	if err != nil {
		return err
	}

	r.tokenManager = tokenmanager.New(fr)

	r.dropboxConfig = dropbox.Config{
		Token:    r.config.Dropbox.Token,
		LogLevel: dropbox.LogInfo, // if needed, set the desired logging level. Default is off
	}
	r.dropboxFiles = files.New(r.dropboxConfig)

	r.fileTracker, err = filetracker.NewFileTracker(r.workDir)
	if err != nil {
		return err
	}

	return nil
}

func (r *App) Close() error {
	if r.tokenManager != nil {
		r.tokenManager.Close()
	}
	if r.fileTracker != nil {
		r.fileTracker.Close()
	}
	return nil
}
