package app

import (
	"strings"
)

type UploadResult string

const (
	UploadResultWait   = UploadResult("wait")
	UploadResultRetry  = UploadResult("retry")
	UploadResultReturn = UploadResult("return")
)

func wrapGoogleError(err error) UploadResult {
	if err == nil {
		return ""
	}
	e := err.Error()
	if strings.Contains(e, "429: Quota exceeded for quota") {
		if strings.Contains(e, "limit 'Write requests per minute per user'") {
			return UploadResultWait
		}
		if strings.Contains(e, "limit 'All requests per day'") {
			return UploadResultReturn
		}
	}

	return ""
}
