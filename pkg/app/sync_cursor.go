package app

func (r *sync) updateCursor(cursor string, hasMore bool) {
	r.Cursor = cursor
	r.HasMore = hasMore
	r.fileTracker.Set("dropbox.cursor", cursor)
}

func (r *sync) getCursor() string {
	return r.fileTracker.Get("dropbox.cursor")
}
