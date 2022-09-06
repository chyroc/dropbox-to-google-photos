package app

import (
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

// == cursor ==

func (r *syncer) cursorKey() string {
	return "dropbox.cursor"
}

func (r *syncer) updateCursor(cursor string, hasMore bool) {
	r.Cursor = cursor
	r.HasMore = hasMore
	r.fileTracker.Set(r.cursorKey(), []byte(cursor))
}

func (r *syncer) getCursor() string {
	return string(r.fileTracker.Get(r.cursorKey()))
}

// == dropbox file exist: hash ==

func (r *syncer) dropboxHash(item iface.FileItem) string {
	return "dropbox.hash:" + item.(*dropboxFileItem).hash
}

func (r *syncer) checkFileExist(item iface.FileItem) bool {
	return len(r.fileTracker.Get(r.dropboxHash(item))) > 0
}

func (r *syncer) setFileExist(item iface.FileItem) {
	r.fileTracker.Set(r.dropboxHash(item), []byte(item.(*dropboxFileItem).hash))
}

func (r *syncer) setFileSkip(item iface.FileItem) {
	r.fileTracker.Set(r.dropboxHash(item), []byte("skip"))
}

// == google photo upload token exist ==

func (r *syncer) googleUploadTokenKey(item iface.FileItem) string {
	return "dropbox-to-google.upload_token:" + item.(*dropboxFileItem).hash
}

func (r *syncer) getUploadToken(item iface.FileItem) string {
	return string(r.fileTracker.Get(r.googleUploadTokenKey(item)))
}

func (r *syncer) setUploadToken(item iface.FileItem, uploadToken string) {
	r.fileTracker.Set(r.googleUploadTokenKey(item), []byte(uploadToken))
}
