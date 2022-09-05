package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

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

func (r *dropboxFileItem) Name() string {
	return r.name
}

func (r *dropboxFileItem) Size() int64 {
	return r.size
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
	}
	return !notImageAndVideoExtensions[strings.ToLower(filepath.Ext(file))]
}
