package app

import (
	"strings"
)

func wrapGoogleError(err error) UploadResult {
	if err == nil {
		return ""
	}
	e := err.Error()
	if strings.Contains(e, "Error 429: Quota exceeded for quota metric 'Write requests' and limit 'Write requests per minute per user'") {
		return UploadResultWait
	}

	return ""
}
