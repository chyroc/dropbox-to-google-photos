package googlephotoclient

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"google.golang.org/api/googleapi"
)

func (r *Client) uploadPart(ctx context.Context, item iface.FileItem) (string, error) {
	offset := r.offsetFromPreviousSession(ctx, item)
	r.log.Debugf("[google] part upload offset for [%s (%s)] is %d", item.Name(), humanSize(item.Size()), offset)
	if offset == 0 {
		return r.createUploadSession(ctx, item)
	}
	return r.resumeUploadSession(ctx, item, offset)
}

func (r *Client) offsetFromPreviousSession(ctx context.Context, item iface.FileItem) int64 {
	if r.uploadSessionUrl(item) == "" {
		return 0
	}
	req, err := http.NewRequest("POST", r.uploadSessionUrl(item), nil)
	if err != nil {
		return 0
	}
	req.Header.Set("Content-Length", "0")
	req.Header.Set("X-Goog-Upload-Command", "query")
	res, err := r.doRequest(ctx, req)
	if err != nil {
		return 0
	}
	defer res.Body.Close()
	return r.offsetFromResponse(res, item)
}

func (r *Client) offsetFromResponse(res *http.Response, item iface.FileItem) int64 {
	if res.Header.Get("X-Goog-Upload-Status") != "active" {
		// Other known statuses "final" and "cancelled" are both considered as already completed.
		// Let's restart the upload from scratch.
		r.offsetStore.Delete(fingerprint(item))
		return 0
	}

	offset, err := strconv.ParseInt(res.Header.Get("X-Goog-Upload-Size-Received"), 10, 64)
	if err == nil && offset > 0 && offset < item.Size() {
		return offset
	}
	r.offsetStore.Delete(fingerprint(item))
	return 0
}

func (r *Client) createUploadSession(ctx context.Context, item iface.FileItem) (string, error) {
	req, err := r.prepareUploadPartRequest(item)
	if err != nil {
		return "", fmt.Errorf("[google] creating upload session: %w", err)
	}

	res, err := r.doRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("[google] creating upload session: %w", err)
	}
	defer res.Body.Close()

	r.storeUploadSession(res, item)

	return r.resumeUploadSession(ctx, item, 0)
}

func (r *Client) storeUploadSession(res *http.Response, item iface.FileItem) {
	if url := res.Header.Get("X-Goog-Upload-URL"); url != "" {
		r.offsetStore.Set(fingerprint(item), []byte(url))
	}
}

func (r *Client) prepareUploadPartRequest(item iface.FileItem) (*http.Request, error) {
	_, size, err := item.OpenSeeker()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", r.googleAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Length", "0")
	req.Header.Set("X-Goog-Upload-Command", "start")
	req.Header.Set("X-Goog-Upload-Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-File-Name", item.Name())
	req.Header.Set("X-Goog-Upload-Protocol", "resumable")
	req.Header.Set("X-Goog-Upload-Raw-Size", fmt.Sprintf("%d", size))

	return req, nil
}

func (r *Client) resumeUploadSession(ctx context.Context, item iface.FileItem, offset int64) (string, error) {
	r.log.Debugf("Resuming upload session for [%s] starting at offset %d", item.Name(), offset)
	req, err := r.prepareResumeUploadRequest(item, offset)
	if err != nil {
		return "", fmt.Errorf("resuming upload session: %w", err)
	}
	res, err := r.doRequest(ctx, req)
	if err != nil {
		r.log.Errorf("Failed to resume session: err=%s", err)
		return "", fmt.Errorf("resuming upload session: %w", err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r.log.Errorf("Failed to read response %s", err)
		return "", fmt.Errorf("resuming upload session: %w", err)
	}
	token := string(b)
	return token, nil
}

func (r *Client) prepareResumeUploadRequest(item iface.FileItem, offset int64) (*http.Request, error) {
	body, size, err := item.OpenSeeker()
	if err != nil {
		return nil, fmt.Errorf("preparing resume upload request: %w", err)
	}
	if _, err := body.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("preparing resume upload request: %w", err)
	}
	req, err := http.NewRequest("POST", r.uploadSessionUrl(item), body)
	if err != nil {
		return nil, fmt.Errorf("preparing resume upload request: %w", err)
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size-offset))
	req.Header.Add("X-Goog-Upload-Command", "upload, finalize")
	req.Header.Set("X-Goog-Upload-Offset", fmt.Sprintf("%d", offset))

	return req, nil
}

// doRequest executes the request call.
// Exactly one of *httpResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *httpResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header.
func (r *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	res, err := r.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Client) uploadSessionUrl(item iface.FileItem) string {
	return string(r.offsetStore.Get(fingerprint(item)))
}

func fingerprint(item iface.FileItem) string {
	if tmp, ok := item.(iface.FingerPrinter); ok {
		return tmp.FingerPrint()
	}
	return fmt.Sprintf("%s|%d", item.Name(), item.Size())
}
