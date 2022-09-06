package app

import (
	"context"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/oauth"
	"github.com/chyroc/dropbox-to-google-photos/pkg/store"
	"golang.org/x/oauth2"
)

func (r *App) TryAuth() error {
	account := r.config.Account
	r.logger.Infof("Authenticating using token for '%s'", account)

	cfg := &oauth.Config{
		ClientID:     r.config.GooglePhotos.ClientID,
		ClientSecret: r.config.GooglePhotos.ClientSecret,
		Logf:         r.logger.Debugf,
	}
	token, _ := r.tokenManager.Get(account)
	var err error

	if token == nil || token.AccessToken == "" {
		token, err = r.tryNewToken(cfg)
	} else {
		token, err = r.tryRefreshToken(cfg, token)
	}
	if err != nil {
		return err
	}

	if err := r.tokenManager.Put(r.config.Account, token); err != nil {
		r.logger.Debugf("[token] Failed to store token into token manager: %s", err)
	}

	googlePhotoHttpClient, err := oauth.Client(context.Background(), cfg, token)
	if err != nil {
		return err
	}

	r.googlePhotoClient, err = googlephotoclient.New(googlePhotoHttpClient,
		store.WrapPrefixStore("google.reuseupload.offseturl", r.fileTracker),
		r.logger)

	return err
}

func (r *App) tryNewToken(cfg *oauth.Config) (*oauth2.Token, error) {
	token, err := oauth.GetToken(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *App) tryRefreshToken(cfg *oauth.Config, token *oauth2.Token) (*oauth2.Token, error) {
	token, err := oauth.RefreshToken(context.Background(), cfg, token)
	if err != nil {
		return nil, err
	}

	r.logger.Donef("Token is valid, expires at %s", token.Expiry)

	return token, nil
}
