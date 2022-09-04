package googlephotoclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func (u *Client) upload(ctx context.Context, fileItem iface.FileItem) (string, error) {
	req, err := u.prepareUploadRequest(fileItem)
	if err != nil {
		return "", err
	}

	u.log.Debugf("[google] Uploading %s (%d kB)", fileItem.Name(), fileItem.Size()/1024)
	res, err := u.client.Do(req)
	if err != nil {
		u.log.Errorf("Error while uploading %s: %s", fileItem, err)
		return "", err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		u.log.Errorf("Error while uploading %s: %s: could not read body: %s", fileItem, res.Status, err)
		return "", err
	}
	body := string(b)

	if res.StatusCode == http.StatusOK {
		return string(body), nil
	}
	return "", fmt.Errorf("got %s: %s", res.Status, body)
}

// prepareUploadRequest returns an HTTP request to upload item.
func (u *Client) prepareUploadRequest(fileItem iface.FileItem) (*http.Request, error) {
	r, size, err := fileItem.Open()
	if err != nil {
		return nil, err
	}

	url := "https://photoslibrary.googleapis.com/v1/uploads"

	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))
	req.Header.Set("X-Goog-Upload-Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-File-Name", fileItem.Name())
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	return req, nil
}
