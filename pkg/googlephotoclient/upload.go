package googlephotoclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

func (r *Client) upload(ctx context.Context, fileItem iface.FileItem) (string, error) {
	req, err := r.prepareUploadRequest(fileItem)
	if err != nil {
		return "", err
	}

	filekey := fmt.Sprintf("%s (%s)", fileItem.Name(), humanSize(fileItem.Size()))

	r.log.Debugf("[google] uploading %s, start", filekey)
	res, err := r.client.Do(req)
	if err != nil {
		r.log.Errorf("[google] uploading %s, do request fail: %s", filekey, err)
		return "", err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r.log.Errorf("[google] uploading %s, read body fail: %s(%s)", filekey, err, res.Status)
		return "", err
	}
	body := string(b)

	if res.StatusCode == http.StatusOK {
		return string(body), nil
	}
	codeResp := new(codeResp)
	_ = json.Unmarshal(b, codeResp)
	if codeResp.Code != 0 {
		err = fmt.Errorf("[google] uploading %s, %d %s", filekey, codeResp.Code, codeResp.Message)
		r.log.Errorf(err.Error())
		return "", err
	}
	return "", fmt.Errorf("[google] uploading %s, fail, got %s: %s", filekey, res.Status, body)
}

func (r *Client) prepareUploadRequest(fileItem iface.FileItem) (*http.Request, error) {
	body, size, err := fileItem.Open()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", r.googleAPI, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))
	req.Header.Set("X-Goog-Upload-Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-File-Name", fileItem.Name())
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	return req, nil
}

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f kB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
	}
}

type codeResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
