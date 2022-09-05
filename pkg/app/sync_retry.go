package app

import (
	"strings"
)

type UploadResult string

const (
	UploadResultWait          = UploadResult("wait")
	UploadResultRetry         = UploadResult("retry")
	UploadResultReactDayLimit = UploadResult("react_day_limit")
	UpdateResultSkip          = UploadResult("skip")
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
			return UploadResultReactDayLimit
		}
	}

	if strings.Contains(e, "Failed: There was an error while trying to create this media item.") {
		return UpdateResultSkip
	}
	if strings.Contains(e, "Payload must not be empty") {
		return UpdateResultSkip
	}

	return ""
}
