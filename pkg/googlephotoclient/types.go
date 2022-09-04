package googlephotoclient

import (
	"errors"
)

var (
	ErrNotFound     = errors.New("media item not found")
	ErrServerFailed = errors.New("internal server error")
)

// MediaItem represents of a media item (such as a photo or video) in Google Photos.
// See: https://developers.google.com/photos/library/reference/rest/v1/mediaItems
type MediaItem struct {
	ID            string
	Description   string
	ProductURL    string
	BaseURL       string
	MimeType      string
	MediaMetadata MediaMetadata
	Filename      string
}

// MediaMetadata represents metadata for a media item.
// See: https://developers.google.com/photos/library/reference/rest/v1/mediaItems
type MediaMetadata struct {
	CreationTime string
	Width        string
	Height       string
}
