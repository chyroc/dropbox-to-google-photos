package googlephotoclient

import (
	"context"
	"net/http"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

type Client struct {
	media  *photosLibraryMediaItemsRepository
	log    iface.Logger
	client *http.Client
}

func New(client *http.Client, logger iface.Logger) (*Client, error) {
	media, err := newPhotosLibraryClient(client)
	if err != nil {
		return nil, err
	}
	return &Client{
		media:  media,
		log:    logger,
		client: client,
	}, nil
}

// UploadFileToLibrary uploads the specified file to Google Photos.
func (c Client) UploadFileToLibrary(ctx context.Context, fileInfo iface.FileItem) (MediaItem, error) {
	token, err := c.upload(ctx, fileInfo)
	if err != nil {
		return MediaItem{}, err
	}

	result, err := c.media.CreateMany(ctx, []string{token})
	if err != nil {
		return MediaItem{}, err
	}
	return result[0], nil
}
