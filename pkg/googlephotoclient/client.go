package googlephotoclient

import (
	"context"
	"net/http"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

type Client struct {
	media       *photosLibraryMediaItemsRepository
	log         iface.Logger
	client      *http.Client
	offsetStore iface.Storer
	googleAPI   string
}

func New(client *http.Client, store iface.Storer, logger iface.Logger) (*Client, error) {
	media, err := newPhotosLibraryClient(client)
	if err != nil {
		return nil, err
	}
	return &Client{
		media:       media,
		log:         logger,
		client:      client,
		offsetStore: store,
		googleAPI:   "https://photoslibrary.googleapis.com/v1/uploads",
	}, nil
}

func (r Client) UploadFile(ctx context.Context, fileInfo iface.FileItem) (string, error) {
	return r.upload(ctx, fileInfo)
}

func (r Client) UploadFilePart(ctx context.Context, fileInfo iface.FileItemSeeker) (string, error) {
	return r.uploadPart(ctx, fileInfo)
}

// UploadFileToLibrary uploads the specified file to Google Photos.
func (r Client) UploadFileToLibrary(ctx context.Context, token string) (MediaItem, error) {
	result, err := r.media.CreateMany(ctx, []string{token})
	if err != nil {
		return MediaItem{}, err
	}
	return result[0], nil
}
