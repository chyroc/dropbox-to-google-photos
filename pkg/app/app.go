package app

import (
	stdlog "log"
	"os"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"github.com/chyroc/dropbox-to-google-photos/pkg/log"
	"github.com/chyroc/dropbox-to-google-photos/pkg/tokenmanager"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/sirupsen/logrus"
)

type App struct {
	workDir    string
	config     *Config
	logger     iface.Logger
	configName string

	tokenManager      *tokenmanager.TokenManager
	googlePhotoClient *googlephotoclient.Client
	dropboxConfig     dropbox.Config
	dropboxFiles      files.Client
	fileTracker       iface.Storer
}

func NewApp(configName string, level logrus.Level) *App {
	if configName == "" {
		configName = "config.json"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		stdlog.Fatalln(err)
		return nil
	}
	workDir := home + "/.dropbox-to-google-photos"
	return &App{
		workDir:    workDir,
		logger:     log.NewStdout(level),
		configName: configName,
	}
}
