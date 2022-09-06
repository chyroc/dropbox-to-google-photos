package app

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func (r *sync) loadDropboxImages() error {
	res, err := r.dropboxFiles.ListFolderContinue(&files.ListFolderContinueArg{
		Cursor: r.Cursor,
	})
	if err != nil {
		return fmt.Errorf("[sync] list folder continue fail: %w", err)
	}

	for _, v := range res.Entries {
		if fi := r.dropboxMetadataToFileItem(v); fi != nil {
			r.logger.Debugf("[sync] append file: '%s', hash: '%s'", fi.Name(), fi.(*dropboxFileItem).hash)
			r.Files <- fi
		}
	}
	r.updateCursor(res.Cursor, res.HasMore)
	return nil
}

func (r *sync) dropboxMetadataToFileItem(fi files.IsMetadata) iface.FileItem {
	switch fi := fi.(type) {
	case *files.FileMetadata:
		if isShouldSync(fi.Name) {
			return newDropboxFileItem(r.dropboxFiles, fi)
		}
		return nil
	default:
		return nil
	}
}

type dropboxFileItem struct {
	dropboxFiles files.Client
	rev          string
	size         int64
	name         string
	hash         string
}

func newDropboxFileItem(dropboxFiles files.Client, file *files.FileMetadata) iface.FileItem {
	if strings.HasPrefix(file.Name, ".") {
		return nil
	}
	return &dropboxFileItem{
		dropboxFiles: dropboxFiles,
		rev:          file.Rev,
		size:         int64(file.Size),
		hash:         file.ContentHash,
		name:         dropboxNameToGooglePhoto(file.PathDisplay),
	}
}

func (r *dropboxFileItem) Open() (io.Reader, int64, error) {
	_, content, err := r.dropboxFiles.Download(&files.DownloadArg{
		Path: "rev:" + r.rev,
	})
	if err != nil {
		return nil, 0, err
	}
	return content, r.size, nil
}

func (r *dropboxFileItem) OpenSeeker() (io.ReadSeekCloser, int64, error) {
	res, err := r.dropboxFiles.GetTemporaryLink(&files.GetTemporaryLinkArg{
		Path: "rev:" + r.rev,
	})
	if err != nil {
		return nil, 0, err
	}
	return &dropboxFileItemSeeker{link: res.Link, size: int64(res.Metadata.Size)}, r.size, nil
}

func (r *dropboxFileItem) Name() string {
	return r.name
}

func (r *dropboxFileItem) Size() int64 {
	return r.size
}

func (r *dropboxFileItem) FingerPrint() string {
	return r.hash
}

type dropboxFileItemSeeker struct {
	link   string
	size   int64
	offset int64
	reader io.ReadCloser
}

func (r *dropboxFileItemSeeker) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		req, err := http.NewRequest(http.MethodGet, r.link, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", r.offset))
		resp, err := downloadHTTPClient.Do(req)
		if err != nil {
			return 0, err
		}
		r.reader = resp.Body
	}

	return r.reader.Read(p)
}

var downloadHTTPClient = &http.Client{
	Timeout: time.Minute * 60,
}

func (r *dropboxFileItemSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.offset = offset
		return r.offset, nil
	default:
		return 0, fmt.Errorf("not support whence: %d", whence)
	}
}

func (r *dropboxFileItemSeeker) Close() error {
	if r.reader == nil {
		return nil
	}
	return r.reader.Close()
}

func dropboxNameToGooglePhoto(name string) string {
	for _, v := range []string{"/", "-", ":", " ", "(", ")", "[", "]", "{", "}", "!", "@", "#", "$", "%", "^", "&", "*", "+", "=", "|", "\\", ",", "?", ";", "'", "\"", "`", "~"} {
		name = strings.ReplaceAll(name, v, "_")
	}
	return strings.TrimLeft(name, "_")
}

func isShouldSync(file string) bool {
	notImageAndVideoExtensions := map[string]bool{
		".xlsx":  true,
		".xls":   true,
		".docx":  true,
		".doc":   true,
		".pptx":  true,
		".ppt":   true,
		".pdf":   true,
		".txt":   true,
		".csv":   true,
		".zip":   true,
		".rar":   true,
		".7z":    true,
		".tar":   true,
		".gz":    true,
		".tgz":   true,
		".bz2":   true,
		".tbz":   true,
		".bz":    true,
		".tbz2":  true,
		".xz":    true,
		".txz":   true,
		".lz":    true,
		".tlz":   true,
		".lzma":  true,
		".tlzma": true,
		".zst":   true,
		".tzst":  true,
		".iso":   true,
		".dmg":   true,
		".exe":   true,
		".msi":   true,
		".apk":   true,
		".ipa":   true,
		".deb":   true,
		".rpm":   true,
		".jar":   true,
		".war":   true,
		".ear":   true,
		".psd":   true,
		".lnk":   true,
	}
	return !notImageAndVideoExtensions[strings.ToLower(filepath.Ext(file))]
}
