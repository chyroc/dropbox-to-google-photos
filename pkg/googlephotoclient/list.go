package googlephotoclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gphotosuploader/googlemirror/api/photoslibrary/v1"
)

func (r *Client) ListMediaItems(size int64, nextToken string) (string, []*MediaItem, error) {
	uri, _ := url.Parse("https://photoslibrary.googleapis.com/v1/mediaItems")
	q := uri.Query()
	if size > 0 {
		q.Add("pageSize", fmt.Sprintf("%d", size))
	}
	if nextToken != "" {
		q.Add("pageToken", nextToken)
	}
	uri.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		return "", nil, err
	}
	res, err := r.doRequest(context.Background(), req)
	if err != nil {
		return "", nil, err
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}
	resp := new(listMediaItemsResp)
	err = json.Unmarshal(bs, resp)
	if err != nil {
		return "", nil, err
	} else if resp.Message != "" {
		return "", nil, fmt.Errorf("list media items fail: %s", resp.Message)
	}

	items := []*MediaItem{}
	for _, v := range resp.MediaItems {
		tmp := toMediaItem(v)
		items = append(items, &tmp)
	}
	return resp.NextPageToken, items, nil
}

type listMediaItemsResp struct {
	Code          int                        `json:"code"`
	Message       string                     `json:"message"`
	MediaItems    []*photoslibrary.MediaItem `json:"mediaItems"`
	NextPageToken string                     `json:"nextPageToken"`
}
