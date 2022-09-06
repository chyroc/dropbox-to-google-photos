package app

import (
	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
)

func (r *App) GetGoogleClient() *googlephotoclient.Client {
	return r.googlePhotoClient
}
