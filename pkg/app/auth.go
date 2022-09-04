package app

import (
	"context"
	"fmt"

	"github.com/chyroc/dropbox-to-google-photos/pkg/googlephotoclient"
	"github.com/chyroc/dropbox-to-google-photos/pkg/oauth"
)

func (r *App) Auth() error {
	ctx := context.Background()
	cfg := &oauth.Config{
		ClientID:     r.config.GooglePhotos.ClientID,
		ClientSecret: r.config.GooglePhotos.ClientSecret,
		Logf:         r.logger.Infof,
	}
	token, err := oauth.GetToken(ctx, cfg)
	if err == nil {
		r.logger.Donef("Successful authentication for account '%s'", r.config.Account)
	}

	err = r.tokenManager.Put(r.config.Account, token)
	if err != nil {
		r.logger.Warnf("Failed to save token for account '%s': %s", r.config.Account, err)
	}

	googlePhotoHttpClient, err := oauth.Client(context.Background(), cfg, token)
	if err != nil {
		return err
	}
	r.googlePhotoHttpClient = googlePhotoHttpClient
	r.googlePhotoClient, err = googlephotoclient.New(r.googlePhotoHttpClient)
	if err != nil {
		return err
	}

	return nil
}

func (r *App) TryAuth() error {
	account := r.config.Account
	r.logger.Infof("Authenticating using token for '%s'", account)

	token, err := r.tokenManager.Get(account)
	if err != nil {
		return fmt.Errorf("unable to retrieve token, have you authenticated before?: %w", err)
	}

	cfg := &oauth.Config{
		ClientID:     r.config.GooglePhotos.ClientID,
		ClientSecret: r.config.GooglePhotos.ClientSecret,
		Logf:         r.logger.Debugf,
	}

	token, err = oauth.RefreshToken(context.Background(), cfg, token)
	if err != nil {
		return err
	}

	r.logger.Donef("Token is valid, expires at %s", token.Expiry)

	if err := r.tokenManager.Put(account, token); err != nil {
		r.logger.Debugf("Failed to store token into token manager: %s", err)
	}

	googlePhotoHttpClient, err := oauth.Client(context.Background(), cfg, token)
	if err != nil {
		return err
	}
	r.googlePhotoHttpClient = googlePhotoHttpClient
	r.googlePhotoClient, err = googlephotoclient.New(r.googlePhotoHttpClient)
	if err != nil {
		return err
	}

	return nil
}