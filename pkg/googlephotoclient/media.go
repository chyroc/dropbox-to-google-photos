package googlephotoclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gphotosuploader/googlemirror/api/photoslibrary/v1"
	"google.golang.org/api/googleapi"
)

// photosLibraryMediaItemsRepository represents a media items Google Photos repository.
type photosLibraryMediaItemsRepository struct {
	service  *photoslibrary.MediaItemsService
	basePath string
}

// newPhotosLibraryClient returns a Repository using PhotosLibrary service.
func newPhotosLibraryClient(authenticatedClient *http.Client) (*photosLibraryMediaItemsRepository, error) {
	return newPhotosLibraryClientWithURL(authenticatedClient, "")
}

// newPhotosLibraryClientWithURL returns a Repository using PhotosLibrary service with a custom URL.
func newPhotosLibraryClientWithURL(authenticatedClient *http.Client, url string) (*photosLibraryMediaItemsRepository, error) {
	s, err := photoslibrary.New(authenticatedClient)
	if err != nil {
		return nil, err
	}
	if url != "" {
		s.BasePath = url
	}
	return &photosLibraryMediaItemsRepository{
		service:  photoslibrary.NewMediaItemsService(s),
		basePath: s.BasePath,
	}, nil
}

// URL returns the repository url.
func (r photosLibraryMediaItemsRepository) URL() string {
	return r.basePath
}

// CreateMany creates one or more media items in the repository.
// By default, the media item(s) will be added to the end of the library.
func (r photosLibraryMediaItemsRepository) CreateMany(ctx context.Context, mediaItems []string) ([]MediaItem, error) {
	return r.CreateManyToAlbum(ctx, "", mediaItems)
}

// CreateManyToAlbum creates one or more media item(s) in the repository.
// If an album id is specified, the media item(s) are also added to the album.
// By default, the media item(s) will be added to the end of the library or album.
func (r photosLibraryMediaItemsRepository) CreateManyToAlbum(ctx context.Context, albumId string, uploadTokens []string) ([]MediaItem, error) {
	newMediaItems := make([]*photoslibrary.NewMediaItem, len(uploadTokens))
	for i, uploadToken := range uploadTokens {
		newMediaItems[i] = &photoslibrary.NewMediaItem{
			SimpleMediaItem: &photoslibrary.SimpleMediaItem{UploadToken: uploadToken},
		}
	}
	req := &photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       albumId,
		NewMediaItems: newMediaItems,
	}
	result, err := r.service.BatchCreate(req).Context(ctx).Do()
	if err != nil {
		return []MediaItem{}, err
	}
	mediaItemsResult := make([]MediaItem, len(result.NewMediaItemResults))
	for i, res := range result.NewMediaItemResults {
		// #54: MediaItem is populated if no errors occurred and the media item was created successfully.
		// If an error occurs res.Status should have more data about the error.
		// @see: https://developers.google.com/photos/library/reference/rest/v1/mediaItems/batchCreate#NewMediaItemResult
		if res.Status.Code != 0 {
			return nil, fmt.Errorf("error creating media item: %s", res.Status.Message)
		}
		if res.MediaItem != nil {
			mediaItemsResult[i] = toMediaItem(res.MediaItem)
		}
	}
	return mediaItemsResult, nil
}

// Get returns the media item specified based on a given media item id.
func (r photosLibraryMediaItemsRepository) Get(ctx context.Context, mediaItemId string) (*MediaItem, error) {
	result, err := r.service.Get(mediaItemId).Context(ctx).Do()
	if err != nil && err.(*googleapi.Error).Code == http.StatusNotFound {
		return &MediaItem{}, ErrNotFound
	}
	if err != nil {
		return &MediaItem{}, ErrServerFailed
	}
	m := toMediaItem(result)
	return &m, nil
}

// maxItemsPerPage is the maximum number of media items to ask to the PhotosLibrary. Fewer media items might
// be returned than the specified number. See https://developers.google.com/photos/library/guides/list#pagination
const maxItemsPerPage = 100

// ListByAlbum list all media items in the specified album.
func (r photosLibraryMediaItemsRepository) ListByAlbum(ctx context.Context, albumId string) ([]MediaItem, error) {
	req := &photoslibrary.SearchMediaItemsRequest{
		AlbumId:  albumId,
		PageSize: maxItemsPerPage,
	}

	photosMediaItems := make([]*photoslibrary.MediaItem, 0)
	appendResultsFn := func(result *photoslibrary.SearchMediaItemsResponse) error {
		photosMediaItems = append(photosMediaItems, result.MediaItems...)
		return nil
	}

	if err := r.service.Search(req).Pages(ctx, appendResultsFn); err != nil {
		return []MediaItem{}, err
	}

	mediaItems := make([]MediaItem, len(photosMediaItems))
	for i, item := range photosMediaItems {
		mediaItems[i] = toMediaItem(item)
	}

	return mediaItems, nil
}

// toMediaItem transforms a `photoslibrary.MediaItem` into a `MediaItem`.
func toMediaItem(item *photoslibrary.MediaItem) MediaItem {
	return MediaItem{
		ID:         item.Id,
		ProductURL: item.ProductUrl,
		BaseURL:    item.BaseUrl,
		MimeType:   item.MimeType,
		MediaMetadata: MediaMetadata{
			CreationTime: item.MediaMetadata.CreationTime,
			Width:        strconv.FormatInt(item.MediaMetadata.Width, 10),
			Height:       strconv.FormatInt(item.MediaMetadata.Height, 10),
		},
		Filename: item.Filename,
	}
}
